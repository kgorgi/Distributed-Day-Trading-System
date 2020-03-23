package main

import (
	"fmt"
	"log"
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

	var cacheSwitch string
	if noCache {
		cacheSwitch = "n"
	} else {
		cacheSwitch = "y"
	}
	payload := fmt.Sprintf("%d,%s,%s,%s,%s", auditClient.TransactionNum, auditClient.Command, stockSymbol, userID, cacheSwitch)

	status, body, err := lib.ClientSendRequest(serverurls.Env.QuoteCacheServer, payload)
	if err != nil {
		log.Fatalln("Connection Error: " + err.Error())
		return 0
	}

	if status != lib.StatusOk {
		log.Fatalln("Response Error: Status " + strconv.Itoa(status) + " " + body)
		return 0
	}

	// Process Data
	quote, err := strconv.ParseUint(body, 10, 64)
	if err != nil {
		log.Fatalln("Received invalid data from quote cache server")
		return 0
	}

	return quote
}
