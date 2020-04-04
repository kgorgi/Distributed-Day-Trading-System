package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"extremeWorkload.com/daytrader/lib"

	"extremeWorkload.com/daytrader/lib/serverurls"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

const dbPoolCount = 100

var (
	ErrNotFound   = errors.New("The specified document does not exist")
	ErrEmptyStack = errors.New("The specified stack is empty")
)

var client *mongo.Client

func InitDatabaseConnection() {
	name, nameOk := os.LookupEnv("USER_NAME")
	pass, passOk := os.LookupEnv("USER_PASS")
	if !nameOk || !passOk {
		log.Fatal("Environment Variables for mongo auth were not set properly")
	}

	//hookup to mongo
	clientOptions := options.Client().ApplyURI(serverurls.Env.DataDBServer).SetAuth(options.Credential{AuthSource: "extremeworkload", Username: name, Password: pass})

	clientOptions.SetMaxPoolSize(dbPoolCount)
	clientOptions.SetMinPoolSize(dbPoolCount)

	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")
}

func CreateTrigger(trigger Trigger) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	_, err := collection.InsertOne(context.TODO(), trigger)
	return err
}

// CheckTriggersIterator returns an iterator function.
// The iterator function returns one trigger from the DB everytime it is
// called, it returns false when all triggers have been returned.
// Note only triggers that have a price set are returned.
func CheckTriggersIterator() (func() (bool, Trigger, error), error) {
	options := options.Find()
	options.SetSort(bson.D{{Key: "stock", Value: -1}, {Key: "_id", Value: -1}})

	query := bson.M{"price_cents": bson.M{"$gt": 0}}
	collection := client.Database("extremeworkload").Collection("triggers")
	cursor, err := collection.Find(context.TODO(), query, options)
	if err != nil {
		return nil, err
	}

	return func() (bool, Trigger, error) {
		if cursor.Next(context.TODO()) {
			var trigger Trigger
			err := cursor.Decode(&trigger)
			if err != nil {
				return false, Trigger{}, err
			}

			return true, trigger, nil
		}

		cursor.Close(context.TODO())
		return false, Trigger{}, nil
	}, nil
}

func ReadTriggersByUser(commandID string) ([]Trigger, error) {
	query := bson.M{"user_command_id": commandID}
	collection := client.Database("extremeworkload").Collection("triggers")
	cursor, err := collection.Find(context.TODO(), query)
	if err == mongo.ErrNoDocuments {
		return []Trigger{}, ErrNotFound
	}

	if err != nil {
		return []Trigger{}, err
	}

	// copy over triggers
	var triggers []Trigger
	for cursor.Next(context.TODO()) {
		var trigger Trigger
		err := cursor.Decode(&trigger)
		if err != nil {
			return []Trigger{}, err
		}

		triggers = append(triggers, trigger)
	}

	cursor.Close(context.TODO())
	return triggers, nil
}

func ReadTrigger(commandID string, stock string, isSell bool) (Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")

	var trigger Trigger
	err := collection.FindOne(context.TODO(), bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}).Decode(&trigger)

	if err == mongo.ErrNoDocuments {
		return trigger, ErrNotFound
	}
	return trigger, err
}

func DeleteTrigger(commandID string, stock string, isSell bool) (Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")
	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}

	var deletedTrigger Trigger
	err := collection.FindOneAndDelete(context.TODO(), filter).Decode(&deletedTrigger)

	if err == mongo.ErrNoDocuments {
		return deletedTrigger, ErrNotFound
	}

	return deletedTrigger, err
}

func CreateUser(user User) error {
	collection := client.Database("extremeworkload").Collection("users")
	_, err := collection.InsertOne(context.TODO(), user)
	return err
}

func ReadUsers() ([]User, error) {
	collection := client.Database("extremeworkload").Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	//copy over users
	var users []User
	for cursor.Next(context.TODO()) {
		var user User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	cursor.Close(context.TODO())
	return users, nil
}

func ReadUser(commandID string) (User, error) {
	collection := client.Database("extremeworkload").Collection("users")

	var user User
	err := collection.FindOne(context.TODO(), bson.M{"command_id": commandID}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return user, ErrNotFound
	}

	return user, err
}

func UpdateUser(userID string, stock string, amount int, cents int, auditClient *auditclient.AuditClient) error {
	// If no stock should be added or removed
	if stock == "" || amount == 0 {
		updateErr := UpdateCents(userID, cents)
		if updateErr != nil {
			return updateErr
		}

		return nil
	}

	updateErr := UpdateStockAndCents(userID, stock, amount, cents)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

// Add a specified amount of stock, and remove a specified amount of cents to a user.
// If a user cannot be found, or they lack sufficent funds or stock, ErrNotFound is returned.
func UpdateStockAndCents(commandID string, stock string, amount int, cents int) error {
	collection := client.Database("extremeworkload").Collection("users")

	var filter bson.M
	var update bson.M
	var err error

	if amount == 0 {
		return errors.New("If you don't want to update stock use the updateCents function")
	}

	// First, if the stock is being added, if the user has no stock add some
	if amount > 0 {
		emptyInvestment := Investment{Stock: stock, Amount: 0}
		filter := bson.M{"command_id": commandID, "investments.stock": bson.M{"$ne": stock}}
		update := bson.M{"$push": bson.M{"investments": emptyInvestment}}
		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}

	// Next, increment the stock by the specified amount making sure the user has enough money and stock
	filter = bson.M{"command_id": commandID, "cents": bson.M{"$gte": -cents}, "investments.stock": stock, "investments.amount": bson.M{"$gte": -amount}}
	update = bson.M{"$inc": bson.M{"investments.$.amount": amount, "cents": cents}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	// If nothing was updated, return an error
	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return ErrNotFound
	}

	// if stock was removed and the user has none left remove the investment from the user
	if amount < 0 {
		// If there's no stock left, remove the investment from the user
		emptyInvestment := Investment{Stock: stock, Amount: 0}
		filter = bson.M{"command_id": commandID, "investments.stock": stock, "investments.amount": 0}
		update = bson.M{"$pull": bson.M{"investments": emptyInvestment}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateCents(commandID string, amount int) error {
	collection := client.Database("extremeworkload").Collection("users")

	filter := bson.M{"command_id": commandID, "cents": bson.M{"$gte": -amount}}
	update := bson.M{"$inc": bson.M{"cents": amount}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func UpdateTriggerPrice(commandID string, stock string, isSell bool, price uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell, "amount_cents": bson.M{"$gte": price}}
	update := bson.M{"$set": bson.M{"price_cents": price}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func UpdateTriggerAmount(commandID string, stock string, isSell bool, amount uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}
	update := bson.M{"$set": bson.M{"amount_cents": amount}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func PushUserReserve(commandID string, stock string, cents uint64, numStocks uint64, isSell bool) error {
	collection := client.Database("extremeworkload").Collection("users")
	reserve := "buys"
	if isSell {
		reserve = "sells"
	}

	newReserve := Reserve{Stock: stock, Cents: cents, Num_Stocks: numStocks, Timestamp: lib.GetUnixTimestamp()}
	filter := bson.M{"command_id": commandID}
	update := bson.M{"$push": bson.M{reserve: bson.M{"$each": bson.A{newReserve}, "$position": 0}}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func PopUserReserve(commandID string, isSell bool) (Reserve, error) {
	collection := client.Database("extremeworkload").Collection("users")
	reserve := "buys"
	if isSell {
		reserve = "sells"
	}

	// delete all buys that are older than 60s
	filter := bson.M{"command_id": commandID}
	update := bson.M{"$pull": bson.M{reserve: bson.M{"timestamp": bson.M{"$lte": lib.GetUnixTimestamp() - (60 * 1000)}}}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return Reserve{}, err
	}

	if result.MatchedCount == 0 {
		return Reserve{}, ErrNotFound
	}

	// remove the front element from the array
	filter = bson.M{"command_id": commandID}
	update = bson.M{"$pop": bson.M{reserve: -1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.Before)

	// get a copy of the user before it was updated
	var user User
	err = collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return Reserve{}, ErrNotFound
	}

	if err != nil {
		return Reserve{}, err
	}

	var reserves []Reserve
	if isSell {
		reserves = user.Sells
	} else {
		reserves = user.Buys
	}

	if len(reserves) == 0 {
		return Reserve{}, ErrEmptyStack
	}

	return reserves[0], nil
}
