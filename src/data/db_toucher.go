package main

import (
	"context"
	"errors"
	"fmt"

	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Better, more detailed error handling.

var (
	errNotFound = errors.New("The specified document does not exist")
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

func readTriggersByUser(client *mongo.Client, commandID string) ([]modelsdata.Trigger, error) {
	triggers, err := queryTriggers(client, bson.M{"user_command_id": commandID})
	if err != nil {
		return []modelsdata.Trigger{}, err
	}

	return triggers, nil
}

func readTrigger(client *mongo.Client, commandID string, stock string, isSell bool) (modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")

	var trigger modelsdata.Trigger
	err := collection.FindOne(context.TODO(), bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}).Decode(&trigger)
	return trigger, err
}

func deleteTrigger(client *mongo.Client, commandID string, stock string, isSell bool) (modelsdata.Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")
	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}

	var deletedTrigger modelsdata.Trigger
	err := collection.FindOneAndDelete(context.TODO(), filter).Decode(&deletedTrigger)

	if err == mongo.ErrNoDocuments {
		return deletedTrigger, errNotFound
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

func readUser(client *mongo.Client, commandID string) (modelsdata.User, error) {
	collection := client.Database("extremeworkload").Collection("users")

	var user modelsdata.User
	err := collection.FindOne(context.TODO(), bson.M{"command_id": commandID}).Decode(&user)
	return user, err
}

func updateUser(client *mongo.Client, user modelsdata.User) error {
	collection := client.Database("extremeworkload").Collection("users")

	update := bson.M{
		"$set": bson.M{"cents": user.Cents, "investments": user.Investments},
	}

	filter := bson.M{"command_id": user.Command_ID}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Add a specified amount of stock, and remove a specified amount of cents to a user.
// If a user cannot be found, or they lack sufficent funds or stock, errNotFound is returned.
func updateStockAndCents(client *mongo.Client, commandID string, stock string, amount int, cents int) error {
	collection := client.Database("extremeworkload").Collection("users")

	var filter bson.M
	var update bson.M
	var err error

	if amount == 0 {
		return errors.New("If you don't want to update stock use the updateCents function")
	}

	// First, if the stock is being added, if the user has no stock add some
	if amount > 0 {
		emptyInvestment := modelsdata.Investment{Stock: stock, Amount: 0}
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
		fmt.Println("Either the user doesn't exist or they do not have sufficient funds or stock")
		return errNotFound
	}

	// if stock was removed and the user has none left remove the investment from the user
	if amount < 0 {
		// If there's no stock left, remove the investment from the user
		emptyInvestment := modelsdata.Investment{Stock: stock, Amount: 0}
		filter = bson.M{"command_id": commandID, "investments.stock": stock, "investments.amount": 0}
		update = bson.M{"$pull": bson.M{"investments": emptyInvestment}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateCents(client *mongo.Client, commandID string, amount int) error {
	collection := client.Database("extremeworkload").Collection("users")

	filter := bson.M{"command_id": commandID, "cents": bson.M{"$gte": -amount}}
	update := bson.M{"$inc": bson.M{"cents": amount}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		fmt.Println("The specified user either does not exist or does not have sufficient funds to remove " + string(amount) + " cents")
		return errNotFound
	}

	return nil
}

func updateTriggerPrice(client *mongo.Client, commandID string, stock string, isSell bool, price uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell, "amount_cents": bson.M{"$gte": price}}
	update := bson.M{"$set": bson.M{"price_cents": price}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return errNotFound
	}

	return nil
}

func updateTriggerAmount(client *mongo.Client, commandID string, stock string, isSell bool, amount uint64) error {
	collection := client.Database("extremeworkload").Collection("triggers")

	filter := bson.M{"user_command_id": commandID, "stock": stock, "is_sell": isSell}
	update := bson.M{"$set": bson.M{"amount_cents": amount}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		return errNotFound
	}

	return nil
}
