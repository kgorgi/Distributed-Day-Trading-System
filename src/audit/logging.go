package main

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"sync/atomic"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

var transactionNum uint64 = 0

func handleLog(conn *net.Conn, payload string) {
	var result interface{}

	err := json.Unmarshal([]byte(payload), &result)

	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	logToConsole(result)

	err = writeToDatabase(result)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(*conn)
}

func handleUserCommand(conn *net.Conn, payload string) {
	// Set Proper Transaction Number
	var internalInfo auditclient.InternalLogInfo
	err := json.Unmarshal([]byte(payload), &internalInfo)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	var userCommandInfo auditclient.UserCommandInfo
	err = json.Unmarshal([]byte(payload), &userCommandInfo)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	nextNum := atomic.AddUint64(&transactionNum, 1)
	internalInfo.TransactionNum = nextNum

	result := struct {
		*auditclient.InternalLogInfo `bson:",inline"`
		*auditclient.UserCommandInfo `bson:",inline"`
	}{
		&internalInfo,
		&userCommandInfo,
	}

	logToConsole(result)

	err = writeToDatabase(result)
	if err != nil {
		lib.ServerSendResponse(*conn, lib.StatusSystemError, err.Error())
		return
	}

	// Send TransactionNumber to Web Server
	lib.ServerSendResponse(*conn, lib.StatusOk, strconv.FormatUint(nextNum, 10))
}

func logToConsole(data interface{}) {
	output, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		lib.Debugln("Unable to output audit message to console: " + err.Error())
		return
	}

	lib.Debugln(string(output))
}

func writeToDatabase(data interface{}) error {
	collection := client.Database("audit").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data)
	return err
}
