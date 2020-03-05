package main

import (
	"context"
	"errors"
	"fmt"

	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound = errors.New("The specified document does not exist")
)

func queryTriggers(client *mongo.Client, query bson.M) ([]modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")
	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	//copy over users
	var triggers []modelsdata.Trigger
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var trigger modelsdata.Trigger
		cursor.Decode(&trigger)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func createTrigger(client *mongo.Client, trigger modelsdata.Trigger) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	_, err := collection.InsertOne(context.TODO(), trigger)
	return err
}

func readTriggers(client *mongo.Client) ([]modelsdata.Trigger, error) {
	triggers, err := queryTriggers(client, bson.M{})
	if err != nil {
		return []modelsdata.Trigger{}, err
	}

	return triggers, nil
}

func readTriggersByUser(client *mongo.Client, user_command_ID string) ([]modelsdata.Trigger, error) {
	triggers, err := queryTriggers(client, bson.M{"user_command_id": user_command_ID})
	if err != nil {
		return []modelsdata.Trigger{}, err
	}

	return triggers, nil
}

func readTrigger(client *mongo.Client, user_command_ID string, stock string, isSell bool) (modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")

	var trigger modelsdata.Trigger
	err := collection.FindOne(context.TODO(), bson.M{"user_command_id": user_command_ID, "stock": stock, "is_sell": isSell}).Decode(&trigger)
	return trigger, err
}

func updateTrigger(client *mongo.Client, trigger modelsdata.Trigger) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	update := bson.D{
		{"$set", bson.M{"price_cents": trigger.Price_Cents, "amount_cents": trigger.Amount_Cents}},
	}
	filter := bson.M{"user_command_id": trigger.User_Command_ID, "stock": trigger.Stock}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func deleteTrigger(client *mongo.Client, user_command_ID string, stock string, isSell bool) (modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")
	filter := bson.M{"user_command_id": user_command_ID, "stock": stock, "is_sell": isSell}

	var deletedTrigger modelsdata.Trigger
	err := collection.FindOneAndDelete(context.TODO(), filter).Decode(&deletedTrigger)

	if err == mongo.ErrNoDocuments {
		return deletedTrigger, ErrNotFound
	}

	return deletedTrigger, err
}

func createUser(client *mongo.Client, user modelsdata.User) error {
	collection := client.Database("extremeworkload").Collection("users")
	_, err := collection.InsertOne(context.TODO(), user)
	return err
}

func readUsers(client *mongo.Client) ([]modelsdata.User, error) {
	collection := client.Database("extremeworkload").Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	//copy over users
	var users []modelsdata.User
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var user modelsdata.User
		cursor.Decode(&user)
		users = append(users, user)
	}

	return users, nil
}

func readUser(client *mongo.Client, command_ID string) (modelsdata.User, error) {
	collection := client.Database("extremeworkload").Collection("users")

	var user modelsdata.User
	err := collection.FindOne(context.TODO(), bson.D{{"command_id", command_ID}}).Decode(&user)
	return user, err
}

func updateUser(client *mongo.Client, user modelsdata.User) error {
	collection := client.Database("extremeworkload").Collection("users")

	update := bson.D{
		{"$set", bson.M{"cents": user.Cents, "investments": user.Investments}},
	}

	filter := bson.M{"command_id": user.Command_ID}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Add a specified amount of stock, and remove a specified amount of cents to a user.
// If a user cannot be found, or they lack sufficent funds, ErrNotFound is returned.
func buyStock(client *mongo.Client, command_ID string, stock string, amount uint64, cents uint64) error {
	collection := client.Database("extremeworkload").Collection("users")

	// First, add the stock with an amount of 0 if the user doesn't have any
	emptyInvestment := modelsdata.Investment{stock, 0}
	filter := bson.M{"command_id": command_ID, "investments.stock": bson.M{"$ne": stock}}
	update := bson.M{"$push": bson.M{"investments": emptyInvestment}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	// Next, increment the stock by the specified amount
	filter = bson.M{"command_id": command_ID, "cents": bson.M{"$gte": cents}, "investments.stock": stock}
	update = bson.M{"$inc": bson.M{"investments.$.amount": amount, "cents": (int(cents) * -1)}}
	result, err = collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	// If nothing was updated, return an error
	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		fmt.Println("Either the user doesn't exist or they do not have sufficient funds")
		return ErrNotFound
	}

	return nil
}

// Remove a specified amount of stock, and add a specified amount of cents from a user.
// If a user cannot be found, or they lack the sufficient amount of stock ErrNotFound
// is returned.
func sellStock(client *mongo.Client, command_ID string, stock string, amount uint64, cents uint64) error {
	collection := client.Database("extremeworkload").Collection("users")

	// First, if the user has an investment that is large enough to remove the specified amount, then remove it.
	filter := bson.M{"command_id": command_ID, "investments.stock": stock, "investments.amount": bson.M{"$gte": amount}}
	update := bson.M{"$inc": bson.M{"investments.$.amount": (int(amount) * -1), "cents": cents}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	// If not, return an error
	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		fmt.Println("The user with id " + command_ID + " doesn't exist or does not have enough stock to remove the amount " + string(amount))
		return ErrNotFound
	}

	// If there's no stock left, remove the investment from the user
	emptyInvestment := modelsdata.Investment{stock, 0}
	filter = bson.M{"command_id": command_ID, "investments.stock": stock, "investments.amount": 0}
	update = bson.M{"$pull": bson.M{"investments": emptyInvestment}}
	result, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Add a specified amount of money to a user
func addAmount(client *mongo.Client, command_ID string, amount uint64) error {
	collection := client.Database("extremeworkload").Collection("users")

	filter := bson.M{"command_id": command_ID}
	update := bson.M{"$inc": bson.M{"cents": amount}}
	_, err := collection.UpdateOne(context.TODO(), filter, update)

	// possibly add upserting here so the check doesn't need to happen else where in the system

	return err
}

// Remove a specified amount of money from a user
func remAmount(client *mongo.Client, command_ID string, amount uint64) error {
	collection := client.Database("extremeworkload").Collection("users")

	filter := bson.M{"command_id": command_ID, "cents": bson.M{"$gte": amount}}
	update := bson.M{"$inc": bson.M{"cents": (int(amount) * -1)}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		fmt.Println("The specified user either does not exist or does not have sufficient funds to remove " + string(amount) + " cents")
		return ErrNotFound
	}

	return nil
}

func updateTriggerPrice(client *mongo.Client, command_ID string, stock string, isSell bool, price uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": command_ID, "stock": stock, "is_sell": isSell, "amount_cents": bson.M{"$gte": price}}
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

func setTriggerAmount(client *mongo.Client, command_ID string, stock string, isSell bool, amount uint64, transaction_number uint64) (*modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": command_ID, "stock": stock, "is_sell": isSell}
	update := bson.M{"$set": bson.M{"amount_cents": amount}, "$setOnInsert": bson.M{"price_cents": 0, "transaction_number": transaction_number}}
	options := options.FindOneAndUpdate().SetUpsert(true)

	var oldTrigger modelsdata.Trigger
	err := collection.FindOneAndUpdate(context.TODO(), filter, update, options).Decode(&oldTrigger)
	if err != nil {

		// If no documents matched, a new document was created and there is no old trigger to return
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return &oldTrigger, nil
}

func updateTriggerAmount(client *mongo.Client, command_ID string, stock string, isSell bool, amount uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": command_ID, "stock": stock, "is_sell": isSell}
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
