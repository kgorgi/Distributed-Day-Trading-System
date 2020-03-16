package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/serverurls"
)

// GetQuote returns a quote from the quote cache server
func GetQuote(
	stockSymbol string,
	userID string,
	noCache bool,
	auditClient *auditclient.AuditClient) uint64 {

	conn, err := net.Dial("tcp", serverurls.Env.QuoteCacheServer)
	if err != nil {
		log.Fatalln("Could not connect to quote server")
		conn.Close()
		return 0
	}

	var cacheSwitch string
	if noCache {
		cacheSwitch = "n"
	} else {
		cacheSwitch = "y"
	}
	payload := fmt.Sprintf("%d,%s,%s,%s,%s", auditClient.TransactionNum, auditClient.Command, stockSymbol, userID, cacheSwitch)

	status, body, err := lib.ClientSendRequest(conn, payload)
	if err != nil {
		log.Fatalln("Connection Error: " + err.Error())
		conn.Close()
		return 0
	}

	if status != lib.StatusOk {
		log.Fatalln("Response Error: Status " + strconv.Itoa(status) + " " + body)
		conn.Close()
		return 0
	}

	// Process Data
	quote, err := strconv.ParseUint(body, 10, 64)
	if err != nil {
		log.Fatalln("Received invalid data from quote cache server")
		conn.Close()
		return 0
	}

	conn.Close()
	return quote
}
