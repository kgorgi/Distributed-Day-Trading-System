package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/perftools"
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

func handleWebConnection(queue chan *perftools.PerfConn) {
	for {
		conn := <-queue

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
		conn.SetAuditClient(&auditClient)

		processCommand(conn, commandJSON, auditClient)

		conn.Close()
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

	_, check := os.LookupEnv("CHECK_TRIGGERS")
	if check {
		go checkTriggers(&auditclient)
	}

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
