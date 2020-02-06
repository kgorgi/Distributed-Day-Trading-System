package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/resolveurl"
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

	fmt.Println("Started Server on Port 5002")

	client, err = connectToMongo()
	if err != nil {
		fmt.Println(err)
		return
	}

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

		switch data[0] {
		case "LOG":
			handleLog(&conn, data[1])
		case "DUMPLOG":
			handleDumpLog(&conn, data[1])
		default:
			lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Audit Command")
		}
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(resolveurl.AuditDBAddress())
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
