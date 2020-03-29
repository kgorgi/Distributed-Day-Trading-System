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
		errorMessage := "Unable to unmarshal JSON: " + err.Error()
		lib.Errorln(errorMessage)
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	logToDebug(result)

	err = writeToDatabase(result)
	if err != nil {
		errorMessage := "Unable to write logs to database " + err.Error()
		logToError(result, errorMessage)
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	serverSendResponseNoError(*conn, lib.StatusOk, "")
}

func handleUserCommand(conn *net.Conn, payload string) {
	// Set Proper Transaction Number
	var internalInfo auditclient.InternalLogInfo
	err := json.Unmarshal([]byte(payload), &internalInfo)
	if err != nil {
		errorMessage := "Unable to unmarshal internal log info JSON: " + err.Error()
		lib.Errorln(errorMessage)
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	var userCommandInfo auditclient.UserCommandInfo
	err = json.Unmarshal([]byte(payload), &userCommandInfo)
	if err != nil {
		errorMessage := "Unable to unmarshal user command info JSON: " + err.Error()
		lib.Errorln(errorMessage)
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
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

	logToDebug(result)

	err = writeToDatabase(result)
	if err != nil {
		errorMessage := "Failed to write logs to database: " + err.Error()
		logToError(result, errorMessage)
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	// Send TransactionNumber to Web Server
	serverSendResponseNoError(*conn, lib.StatusOk, strconv.FormatUint(nextNum, 10))
}

func logToDebug(data interface{}) {
	output, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		lib.Errorln("Unable to output audit message to console: " + err.Error())
		return
	}

	lib.Debugln(string(output))
}

func logToError(data interface{}, errMessage string) {
	output, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		lib.Errorln("Unable to output audit message to error: " + err.Error())
		return
	}

	fullErrorMessage := "Error: " + errMessage + " Logs: " + string(output)
	lib.Errorln(fullErrorMessage)
}

func writeToDatabase(data interface{}) error {
	collection := client.Database("audit").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data)
	return err
}
