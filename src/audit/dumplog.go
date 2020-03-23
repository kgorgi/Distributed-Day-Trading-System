package main

import (
	"context"
	"fmt"
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
	cursor, _ := collection.Find(context.TODO(), query)

	response := createLog(cursor)
	err := lib.ServerSendResponse(*conn, lib.StatusOk, response)
	if err != nil {
		fmt.Println("Communication Error: " + err.Error())
	}
}

func createLog(cursor *mongo.Cursor) string {
	var str strings.Builder

	str.WriteString("<?xml version=\"1.0\"?>\n")
	str.WriteString("<log>\n")

	for cursor.Next(context.TODO()) {
		var internalInfo auditclient.InternalLogInfo
		cursor.Decode(&internalInfo)

		str.WriteString("\t")
		writeTagHead(&str, internalInfo.LogType)
		str.WriteString("\n")

		logCommand := !(internalInfo.LogType == "accountTransaction" || internalInfo.LogType == "quoteServer")
		writeInternalLogInfoTags(&str, internalInfo, logCommand)
		switch internalInfo.LogType {
		case "userCommand":
			var userLog auditclient.UserCommandInfo
			cursor.Decode(&userLog)
			writeUserCommandTags(&str, userLog)
		case "quoteServer":
			var quoteLog auditclient.QuoteServerResponseInfo
			cursor.Decode(&quoteLog)
			writeQuoteServerTags(&str, quoteLog)
		case "accountTransaction":
			var accountLog auditclient.AccountTransactionInfo
			cursor.Decode(&accountLog)
			writeAccountTransactionTags(&str, accountLog)
		case "errorEvent":
			var errorEvent auditclient.ErrorEventInfo
			cursor.Decode(&errorEvent)
			writeErrorEventTags(&str, errorEvent)
		case "debugEvent":
			var debugEvent auditclient.DebugEventInfo
			cursor.Decode(&debugEvent)
			writeDebugEventTags(&str, debugEvent)
		case "perfMetric":
			var perfLog auditclient.PerformanceMetricInfo
			cursor.Decode(&perfLog)
			writePerfMetricTags(&str, perfLog)
		}

		str.WriteString("\t")
		writeTagTail(&str, internalInfo.LogType)
		str.WriteString("\n")
	}

	str.WriteString("</log>")
	return str.String()
}
