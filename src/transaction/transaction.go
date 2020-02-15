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

func handleWebConnection(conn net.Conn) {
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

	fmt.Println("closed client")
}

func main() {
	fmt.Println("Establishing Database Connection")

	var auditclient = auditclient.AuditClient{
		Server:         "transaction",
		TransactionNum: 0,
		Command:        "N,/A",
	}

	go checkTriggers(auditclient)

	initParameterMaps()
	fmt.Println("Database Server Connected")

	ln, _ := net.Listen("tcp", ":5000")

	for {
		conn, _ := ln.Accept()
		go handleWebConnection(conn)
	}
}
