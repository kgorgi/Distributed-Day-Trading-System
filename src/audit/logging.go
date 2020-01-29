package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"extremeWorkload.com/daytrader/lib"
)

func handleLog(conn *net.Conn, payload string) {
	var result interface{}

	err := json.Unmarshal([]byte(payload), &result)

	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	logToConsole(result)

	collection := client.Database("audit").Collection("logs")
	_, err = collection.InsertOne(context.TODO(), result)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(*conn)
}

func logToConsole(data interface{}) {
	output, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Unable to output audit message to console: " + err.Error())
		return
	}

	fmt.Println(string(output))
}
