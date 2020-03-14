package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/serverurls"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbPoolCount = 500

var auditClient = auditclient.AuditClient{
	Server: "database",
}

func handleConnection(conn net.Conn, client *mongo.Client) {
	lib.Debugln("Connection Established")
	payload, err := lib.ServerReceiveRequest(conn)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	processCommand(conn, client, payload)
	conn.Close()

	lib.Debugln("Connection Closed")
}

func setupIndexes(client *mongo.Client) {

	// User Collection
	userCol := client.Database("extremeworkload").Collection("users")
	mod := mongo.IndexModel{
		Keys: bson.M{
			"command_id": 1, // index in ascending order
		}, Options: nil,
	}

	_, err := userCol.Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		log.Fatal(err)
	}

	// Trigger Indexes
	triggerCol := client.Database("extremeworkload").Collection("triggers")
	mod = mongo.IndexModel{
		Keys: bson.M{
			"user_command_id": 1,
			"stock":           1,
			"is_sell":         1,
		}, Options: nil,
	}

	_, err = triggerCol.Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("Starting data server...")

	//hookup to mongo
	clientOptions := options.Client().ApplyURI(serverurls.Env.DataDBServer)
	clientOptions.SetMaxPoolSize(dbPoolCount)
	clientOptions.SetMinPoolSize(dbPoolCount)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	setupIndexes(client)

	//start listening on the port
	ln, err := net.Listen("tcp", ":5001")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Started data server on port: 5001")

	//connection handling
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}

		go handleConnection(conn, client)
	}
}
