package main

import (
	"context"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleDumpLog(conn *net.Conn, userID string) {
	var query bson.M
	if userID == "" {
		query = bson.M{}
	} else {
		query = bson.M{
			"username": userID,
		}
	}

	collection := client.Database("audit").Collection("logs")
	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		errorMessage := "Failed to find logs: " + err.Error()
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	response, err := createLog(cursor)
	if err != nil {
		errorMessage := "Failed to generate logs: " + err.Error()
		serverSendResponseNoError(*conn, lib.StatusSystemError, errorMessage)
		return
	}

	serverSendResponseNoError(*conn, lib.StatusOk, response)
}

func createLog(cursor *mongo.Cursor) (string, error) {
	var str strings.Builder

	str.WriteString("<?xml version=\"1.0\"?>\n")
	str.WriteString("<log>\n")

	for cursor.Next(context.TODO()) {
		var internalInfo auditclient.InternalLogInfo
		err := cursor.Decode(&internalInfo)
		if err != nil {
			return "", err
		}

		str.WriteString("\t")
		writeTagHead(&str, internalInfo.LogType)
		str.WriteString("\n")

		logCommand := !(internalInfo.LogType == "accountTransaction" || internalInfo.LogType == "quoteServer")
		writeInternalLogInfoTags(&str, internalInfo, logCommand)
		switch internalInfo.LogType {
		case "userCommand":
			var userLog auditclient.UserCommandInfo
			err := cursor.Decode(&userLog)
			if err != nil {
				return "", err
			}
			writeUserCommandTags(&str, userLog)
		case "quoteServer":
			var quoteLog auditclient.QuoteServerResponseInfo
			err := cursor.Decode(&quoteLog)
			if err != nil {
				return "", err
			}
			writeQuoteServerTags(&str, quoteLog)
		case "accountTransaction":
			var accountLog auditclient.AccountTransactionInfo
			err := cursor.Decode(&accountLog)
			if err != nil {
				return "", err
			}
			writeAccountTransactionTags(&str, accountLog)
		case "errorEvent":
			var errorEvent auditclient.ErrorEventInfo
			err := cursor.Decode(&errorEvent)
			if err != nil {
				return "", err
			}
			writeErrorEventTags(&str, errorEvent)
		case "debugEvent":
			var debugEvent auditclient.DebugEventInfo
			err := cursor.Decode(&debugEvent)
			if err != nil {
				return "", err
			}
			writeDebugEventTags(&str, debugEvent)
		case "perfMetric":
			var perfLog auditclient.PerformanceMetricInfo
			err := cursor.Decode(&perfLog)
			if err != nil {
				return "", err
			}
			writePerfMetricTags(&str, perfLog)
		}

		str.WriteString("\t")
		writeTagTail(&str, internalInfo.LogType)
		str.WriteString("\n")
	}

	str.WriteString("</log>")
	cursor.Close(context.TODO())
	return str.String(), nil
}
