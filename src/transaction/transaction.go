package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/security"

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

func handleWebConnection(queue chan net.Conn) {
	for {
		conn := <-queue
		lib.Debugln("Handling Request")

		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			conn.Close()
			return
		}

		var commandJSON CommandJSON
		err = json.Unmarshal([]byte(payload), &commandJSON)
		if err != nil {
			conn.Close()
			return
		}

		transactionNum, _ := strconv.ParseUint(commandJSON.TransactionNum, 10, 64)

		var auditClient = auditclient.AuditClient{
			Server:         "transaction",
			Command:        commandJSON.Command,
			TransactionNum: transactionNum,
		}

		processCommand(conn, commandJSON, auditClient)

		conn.Close()
		lib.Debugln("Connection Closed")
	}

}

func main() {
	fmt.Println("Starting transaction server...")
	security.InitCryptoKey()

	var auditclient = auditclient.AuditClient{
		Server:         "transaction",
		TransactionNum: 0,
		Command:        "",
	}

	go checkTriggers(auditclient)

	ln, _ := net.Listen("tcp", ":5000")
	fmt.Println("Started transaction server on port: 5000")

	queue := make(chan net.Conn, threadCount*10)

	for i := 0; i < threadCount; i++ {
		go handleWebConnection(queue)
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			queue <- conn
		}
	}
}
