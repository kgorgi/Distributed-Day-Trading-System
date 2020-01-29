package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

var auditClient = auditclient.AuditClient{
	Server: "database",
}

type Investment struct {
	Stock  string
	Amount int
}

type User struct {
	Command_ID  string
	Cents       int
	Investments []Investment
}

type Trigger struct {
	User_Command_ID string
	Stock           string
	Price_Cents     int
	Amount_Cents    int
	isSell          bool
}

func readTriggers(client *mongo.Client) ([]Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	//copy over users
	var triggers []Trigger
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var trigger Trigger
		cursor.Decode(&trigger)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func createTrigger(client *mongo.Client, trigger Trigger) {
	collection := client.Database("extremeworkload").Collection("triggers")
	_, err := collection.InsertOne(context.TODO(), trigger)
	if err != nil {
		log.Fatal(err)
	}
}

func readTrigger(client *mongo.Client, user_command_ID string, stock string) (Trigger, error) {
	collection := client.Database("extremeworkload").Collection("triggers")

	var trigger Trigger
	err := collection.FindOne(context.TODO(), bson.M{"user_command_id": user_command_ID, "stock": stock}).Decode(&trigger)
	return trigger, err
}

func updateTrigger(client *mongo.Client, trigger Trigger) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	update := bson.D{
		{"$set", bson.M{"price_cents": trigger.Price_Cents, "amount_cents": trigger.Amount_Cents}},
	}
	filter := bson.M{"user_command_id": trigger.User_Command_ID, "stock": trigger.Stock}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func deleteTrigger(client *mongo.Client, user_command_ID string, stock string) error {
	collection := client.Database("extremeworkload").Collection("triggers")
	filter := bson.M{"user_command_id": user_command_ID, "stock": stock}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}

func readUsers(client *mongo.Client) ([]User, error) {
	collection := client.Database("extremeworkload").Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	//copy over users
	var users []User
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var user User
		cursor.Decode(&user)
		users = append(users, user)
	}

	return users, nil
}

func createUser(client *mongo.Client, user User) error {
	collection := client.Database("extremeworkload").Collection("users")
	_, err := collection.InsertOne(context.TODO(), user)
	return err
}

func readUser(client *mongo.Client, command_ID string) (User, error) {
	collection := client.Database("extremeworkload").Collection("users")

	var user User
	err := collection.FindOne(context.TODO(), bson.D{{"command_id", command_ID}}).Decode(&user)
	return user, err
}

func updateUser(client *mongo.Client, user User) error {
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

func handleConnection(conn net.Conn, client *mongo.Client) {
	for {
		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
			return
		}
		data := strings.Split(payload, "|")
		switch data[0] {
		case "CREATE_USER":
			userJson := data[1]
			var newUser User
			jsonError := json.Unmarshal([]byte(userJson), &newUser)
			if jsonError != nil {
				lib.ServerSendResponse(conn, 400, "json input is incorrect")
				break
			}

			createError := createUser(client, newUser)
			if createError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong")
				break
			}

			lib.ServerSendResponse(conn, 200, "everythings good my dude")

		case "READ_USER":
			commandID := data[1]
			user, readError := readUser(client, commandID)
			userBytes, jsonError := json.Marshal(user)

			if readError != nil && jsonError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong")
			}

			userString := string(userBytes)
			lib.ServerSendResponse(conn, 200, userString)

		case "READ_USERS":
			users, readError := readUsers(client)
			usersBytes, jsonError := json.Marshal(users)

			if readError != nil || jsonError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong")
				break
			}

			usersString := string(usersBytes)
			lib.ServerSendResponse(conn, 200, usersString)

		case "UPDATE_USER":
			userJson := data[1]
			var userUpdate User
			jsonError := json.Unmarshal([]byte(userJson), &userUpdate)
			if jsonError != nil {
				lib.ServerSendResponse(conn, 400, "json input is incorrect")
				break
			}

			updateError := updateUser(client, userUpdate)
			if updateError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong")
				break
			}

			lib.ServerSendResponse(conn, 200, "everythings good my dude")

		case "DELETE_USER":

		case "CREATE_TRIGGER":

		case "READ_TRIGGER":

		case "READ_TRIGGERS":

		case "UPDATE_TRIGGER":

		case "DELETE_TRIGGER":

		default:
			lib.ServerSendResponse(conn, 400, "Invalid Data Server Command")

		}
		lib.ServerSendOKResponse(conn)
	}
}

func main() {
	fmt.Println("Starting Data server...")

	//hookup to mongo
	clientOptions := options.Client().ApplyURI("mongodb://data-mongoDB:27017/mongodb")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	//start listening on the port
	ln, err := net.Listen("tcp", ":5001")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Started Server on Port 5001")

	//connection handling
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Connection Established")
		go handleConnection(conn, client)
	}

	// //create a new user
	// userToCreate := User{"testCommandId",1738, []Investment{}}
	// createUser(client, userToCreate);

	// //update a user
	// var investments []Investment
	// investments = append(investments, Investment{"XXX", 58})
	// userUpdate := User{"testCommandId", 22222, investments}
	// updateUser(client, userUpdate);

	// //find a single user
	// user := readUser(client, "testCommandId")
	// fmt.Println(user);

	// //delete a single user
	// deleteUser(client, "testCommandId");

	// //grab all users
	// users := readUsers(client)
	// fmt.Println(users);

	// //create a trigger
	// triggerToCreate := Trigger{"testCommandId", "ABC", 100, 200, false}
	// createTrigger(client, triggerToCreate);

	// //find a single trigger
	// trigger := readTrigger(client, "testCommandId", "ABC");
	// fmt.Println(trigger)

	// //update a trigger
	// triggerUpdate := Trigger{"testCommandId", "ABC", 333, 222, false}
	// updateTrigger(client, triggerUpdate);

	// //delete a trigger
	// deleteTrigger(client, "testCommandId", "DDD");

	// triggers := readTriggers(client)
	// fmt.Println(triggers)
}
