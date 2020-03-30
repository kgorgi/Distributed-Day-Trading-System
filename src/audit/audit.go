package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
	"extremeWorkload.com/daytrader/lib/serverurls"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbPoolCount = 100

var client *mongo.Client

const threadCount = 1000

func main() {
	fmt.Println("Starting audit server...")
	security.InitCryptoKey()

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

	queue := make(chan net.Conn, threadCount*10)

	for i := 0; i < threadCount; i++ {
		go handleConnection(queue)
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			queue <- conn
		}
	}
}

func handleConnection(queue chan net.Conn) {
	for {
		conn := <-queue

		isHealthCheck, err := lib.ServerReceiveHealthCheck(conn)
		if isHealthCheck || err != nil {
			if err != nil {
				lib.Errorln("Healthcheck Failed: " + err.Error())
			}
			conn.Close()
			continue
		}

		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			lib.Errorln("Failed to receive request: " + err.Error())
			conn.Close()
			continue
		}

		data := strings.Split(payload, "|")

		switch data[0] {
		case "USERLOG":
			handleUserCommand(&conn, data[1])
		case "LOG":
			handleLog(&conn, data[1])
		case "DUMPLOG":
			handleDumpLog(&conn, data[1])
		default:
			serverSendResponseNoError(conn, lib.StatusUserError, "Invalid Audit Command")
		}

		conn.Close()
	}
}

func connectToMongo() (*mongo.Client, error) {
	name, nameOk := os.LookupEnv("USER_NAME")
	pass, passOk := os.LookupEnv("USER_PASS")
	if !nameOk || !passOk {
		return nil, errors.New("Environment Variables for mongo auth were not set properly")
	}

	clientOptions := options.Client().ApplyURI(serverurls.Env.AuditDBServer).SetAuth(options.Credential{AuthSource: "audit", Username: name, Password: pass})
	clientOptions.SetMaxPoolSize(dbPoolCount)
	clientOptions.SetMinPoolSize(dbPoolCount)

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

func serverSendResponseNoError(conn net.Conn, status int, message string) {
	err := lib.ServerSendResponse(conn, status, message)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to send response to %s. %d: %s", conn.RemoteAddr().String(), status, message)
		lib.Errorln(errorMessage)
	}
}
