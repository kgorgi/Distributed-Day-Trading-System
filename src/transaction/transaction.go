package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/perftools"
	"extremeWorkload.com/daytrader/lib/security"
	"extremeWorkload.com/daytrader/transaction/data"

	"extremeWorkload.com/daytrader/lib"
)

type CommandJSON struct {
	TransactionNum string
	Command        string
	Userid         string
	Amount         string
	StockSymbol    string
}

const threadCount = 1000

func handleWebConnection(queue chan *perftools.PerfConn) {
	for {
		conn := <-queue

		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			lib.Errorln("Failed to receive request: " + err.Error())
			conn.Close()
			continue
		}

		var commandJSON CommandJSON
		err = json.Unmarshal([]byte(payload), &commandJSON)
		if err != nil {
			errorMessage := "Failed to unmarshal JSON: " + err.Error()
			lib.Errorln(errorMessage)
			serverSendResponseNoError(conn, lib.StatusSystemError, errorMessage, nil)
			conn.Close()
			continue
		}

		transactionNum, err := strconv.ParseUint(commandJSON.TransactionNum, 10, 64)
		if err != nil {
			errorMessage := "Failed to parse transaction number: " + err.Error()
			lib.Errorln(errorMessage)
			serverSendResponseNoError(conn, lib.StatusSystemError, errorMessage, nil)
			conn.Close()
			continue
		}

		var auditClient = auditclient.AuditClient{
			Server:         "transaction",
			Command:        commandJSON.Command,
			TransactionNum: transactionNum,
		}

		conn.SetAuditClient(&auditClient)

		processCommand(conn, commandJSON, auditClient)

		conn.Close()
	}

}

func main() {
	fmt.Println("Starting transaction server...")
	security.InitCryptoKey()

	data.InitDatabaseConnection(true)

	var auditclient = auditclient.AuditClient{
		Server:         "transaction",
		TransactionNum: 0,
		Command:        "",
	}

	go checkTriggers(&auditclient)

	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Started transaction server on port: 5000")

	queue := make(chan *perftools.PerfConn, threadCount*10)

	for i := 0; i < threadCount; i++ {
		go handleWebConnection(queue)
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			queue <- perftools.NewPerfConn(conn)
		}
	}
}

// ServerSendResponseNoError sends a response to a client and logs all error messages
func serverSendResponseNoError(conn net.Conn, status int, message string, auditClient *auditclient.AuditClient) {
	err := lib.ServerSendResponse(conn, status, message)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to send response to %s. %d: %s", conn.RemoteAddr().String(), status, message)
		lib.Errorln(errorMessage)

		if auditClient != nil {
			auditClient.LogErrorEvent(errorMessage)
		}
	}
}
