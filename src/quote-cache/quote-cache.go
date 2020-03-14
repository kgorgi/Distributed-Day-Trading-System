package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

func handleConnection(conn net.Conn) {
	lib.Debugln("Handling Connection")

	payload, err := lib.ServerReceiveRequest(conn)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		conn.Close()
		return
	}

	// <transaction number>,<command>,<stock symbol>,<userid>,<c,n|cache or no-cache>
	data := strings.Split(payload, ",")

	transactionNum, err := strconv.ParseUint(data[0], 10, 64)
	var auditClient = auditclient.AuditClient{
		Server:         "quote-cache",
		TransactionNum: transactionNum,
		Command: 		data[1],
	}

	var noCache bool

	if data[4] == "n" {
		noCache = true
	} else {
		noCache = false
	}

	quoteVal := GetQuote(data[2], data[3], noCache, &auditClient)
	lib.ServerSendResponse(conn, lib.StatusOk, strconv.FormatUint(quoteVal, 10))

	conn.Close()
	lib.Debugln("Connection Closed")
}

func main() {
	ln, _ := net.Listen("tcp", ":5004")
	fmt.Println("Started quote cache server on port: 5004")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}
