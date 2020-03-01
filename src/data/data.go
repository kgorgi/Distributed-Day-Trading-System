package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/resolveurl"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func main() {
	fmt.Println("Starting data server...")

	//hookup to mongo
	clientOptions := options.Client().ApplyURI(resolveurl.DatabaseDBAddress())
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
