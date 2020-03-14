package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/serverurls"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
	fmt.Println("Starting audit server...")

	ln, err := net.Listen("tcp", ":5002")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Started audit server on port: 5002")

	client, err = connectToMongo()
	if err != nil {
		fmt.Println(err)
		return
	}

	setupIndexes(client)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	lib.Debugln("Connection Established")

	payload, err := lib.ServerReceiveRequest(conn)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		conn.Close()
		return
	}

	data := strings.Split(payload, "|")

	switch data[0] {
	case "USERCOMMAND":
		handleUserCommand(&conn, data[1])
	case "LOG":
		handleLog(&conn, data[1])
	case "DUMPLOG":
		handleDumpLog(&conn, data[1])
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Audit Command")
	}

	conn.Close()
	lib.Debugln("Connection Closed")
}

func setupIndexes(client *mongo.Client) {
	logsCol := client.Database("audit").Collection("logs")
	mod := mongo.IndexModel{
		Keys: bson.M{
			"userID":         1, // index in ascending order
			"timestamp":      1, // index in ascending order
			"transactionNum": 1, // index in ascending order
		}, Options: nil,
	}

	_, err := logsCol.Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		log.Fatal(err)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(serverurls.Env.AuditDBServer)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB")

	return client, nil
}
