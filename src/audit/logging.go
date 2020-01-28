package main

import (
	"context"
	"encoding/json"
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

	collection := client.Database("audit").Collection("logs")
	_, err = collection.InsertOne(context.TODO(), result)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(*conn)
}
