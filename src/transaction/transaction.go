package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	auditclient "extremeWorkload.com/daytrader/lib/audit"

	"extremeWorkload.com/daytrader/lib"
)

type CommandJSON struct {
	TransactionNum string
	Command        string
	Userid         string
	Amount         string
	StockSymbol    string
}

var dataConn databaseWrapper

func handleWebConnection(conn net.Conn) {
	lib.Debugln("Connection Established")

	for {
		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			conn.Close()
			break
		}

		var commandJSON CommandJSON
		err = json.Unmarshal([]byte(payload), &commandJSON)
		if err != nil {
			conn.Close()
			break
		}

		transactionNum, _ := strconv.ParseUint(commandJSON.TransactionNum, 10, 64)

		var auditClient = auditclient.AuditClient{
			Server:         "transaction",
			Command:        commandJSON.Command,
			TransactionNum: transactionNum,
		}

		processCommand(conn, commandJSON, auditClient)
	}

	lib.Debugln("Connection Closed")
}

func main() {
	fmt.Println("Starting transaction server...")

	var auditclient = auditclient.AuditClient{
		Server:         "transaction",
		TransactionNum: 0,
		Command:        "N,/A",
	}

	go checkTriggers(auditclient)

	initParameterMaps()

	ln, _ := net.Listen("tcp", ":5000")
	fmt.Println("Started transaction server on port: 5000")

	for {
		conn, _ := ln.Accept()
		go handleWebConnection(conn)
	}
}
