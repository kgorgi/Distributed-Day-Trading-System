package main

import (
	"encoding/json"
	"fmt"
	"net"

	auditclient "extremeWorkload.com/daytrader/lib/audit"

	"extremeWorkload.com/daytrader/lib"
)

type CommandJSON struct {
	TransactionNum uint64
	Command        string
	Userid         string
	Amount         string
	StockSymbol    string
}

var dataConn databaseWrapper

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

		var auditClient = auditclient.AuditClient{
			Server:         "audit",
			Command:        commandJSON.Command,
			TransactionNum: 1,
		}

		processCommand(conn, commandJSON, auditClient)
	}

	fmt.Println("closed client")
}

func main() {
	fmt.Println("Establishing Database Connection")

	// var err error
	// dataConn.client, err = net.Dial("tcp", "data-server:5001")
	// if err != nil {
	// 	return
	// }
	initParameterMaps()
	fmt.Println("Database Server Connected")

	ln, _ := net.Listen("tcp", ":5000")

	for {
		conn, _ := ln.Accept()
		go handleWebConnection(conn)
	}
}
