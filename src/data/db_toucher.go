package main

import ( 
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "extremeWorkload.com/daytrader/lib/models/data"
);

func createTrigger(client *mongo.Client, trigger modelsdata.Trigger) error{
    collection := client.Database("extremeworkload").Collection("triggers")
    _, err := collection.InsertOne(context.TODO(), trigger);
    return err
}

func readTriggers(client *mongo.Client) ([]modelsdata.Trigger, error) {
    collection := client.Database("extremeworkload").Collection("triggers")
    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        return nil, err
    }

    //copy over users
    var triggers []modelsdata.Trigger
    defer cursor.Close(context.TODO())
    for cursor.Next(context.TODO()) {
        var trigger modelsdata.Trigger
        cursor.Decode(&trigger)
        triggers = append(triggers, trigger);
    }

    return triggers, nil
}

func readTrigger(client *mongo.Client, user_command_ID string, stock string, isSell bool) (modelsdata.Trigger, error) {
    collection := client.Database("extremeworkload").Collection("triggers")

    var trigger modelsdata.Trigger
    err := collection.FindOne(context.TODO(), bson.M{"user_command_id": user_command_ID, "stock": stock, "is_sell": isSell}).Decode(&trigger);
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

func deleteTrigger(client *mongo.Client, user_command_ID string, stock string, isSell bool) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	filter := bson.M{"user_command_id": user_command_ID, "stock": stock, "is_sell": isSell}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}

func createUser(client *mongo.Client, user  modelsdata.User) error{
    collection := client.Database("extremeworkload").Collection("users")
    _, err := collection.InsertOne(context.TODO(), user);
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
        users = append(users, user);
    }

    return users, nil
}

func readUser(client *mongo.Client, command_ID string) (modelsdata.User, error) {
    collection := client.Database("extremeworkload").Collection("users")

    var user modelsdata.User
    err := collection.FindOne(context.TODO(), bson.D{{"command_id", command_ID}}).Decode(&user);
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

func deleteUser(client *mongo.Client, command_ID string) error {
	collection := client.Database("extremeworkload").Collection("users")
	filter := bson.M{"command_id": command_ID}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}