package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/perftools"
	"extremeWorkload.com/daytrader/lib/security"
)

const threadCount = 1000

func handleConnection(queue chan *perftools.PerfConn) {
	for {
		conn := <-queue
		lib.Debugln("Handling Request")

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
			Command:        data[1],
		}
		conn.SetAuditClient(&auditClient)

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
}

func main() {
	ln, err := net.Listen("tcp", ":5004")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Started quote cache server on port: 5004")
	security.InitCryptoKey()

	queue := make(chan *perftools.PerfConn, threadCount*10)

	for i := 0; i < threadCount; i++ {
		go handleConnection(queue)
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			queue <- perftools.NewPerfConn(conn)
		}
	}
}
