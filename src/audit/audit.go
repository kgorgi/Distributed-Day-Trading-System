package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

var client *mongo.Client

func main() {
	fmt.Println("Starting audit server...")

	var err error
	client, err = connectToMongo()
	if err != nil {
		fmt.Println(err)
		return
	}

	ln, err := net.Listen("tcp", ":5002")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Started Server on Port 5002")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Connection Established")

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	for {
		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
			return
		}

		data := strings.Split(payload, "|")

		if data[0] == "LOG" {
			err := handleLog(data[1])
			if err != nil {
				lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
				return
			}
			lib.ServerSendOKResponse(conn)
			return
		} else if data[0] == "DUMPLOG" {
			var logs string
			var err error
			if len(data) == 3 {
				logs, err = handleDumpLog(data[2])
			} else {
				logs, err = handleDumpLog("")
			}

			if err != nil {
				lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
				return
			}

			lib.ServerSendResponse(conn, lib.StatusOk, logs)
			return
		} else {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Audit Command")
			return
		}

	}

}

func handleLog(payload string) error {
	var result interface{}

	err := json.Unmarshal([]byte(payload), &result)
	if err != nil {
		return err
	}

	collection := client.Database("audit").Collection("logs")
	_, err = collection.InsertOne(context.TODO(), result)
	if err != nil {
		return err
	}

	return nil
}

func handleDumpLog(userID string) (string, error) {
	collection := client.Database("audit").Collection("logs")
	cursor, _ := collection.Find(context.TODO(), bson.D{})

	var results []auditclient.InternalLogInfo
	cursor.All(context.TODO(), &results)

	var builder strings.Builder

	for _, element := range results {
		fmt.Fprintln(&builder, element)
	}

	return builder.String(), nil
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://audit-mongoDB:27017/mongodb")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to MongoDB!")

	return client, nil
}
