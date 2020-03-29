package main

import (
	"errors"
	"fmt"
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
	auditClient *auditclient.AuditClient) (uint64, error) {

	var cacheSwitch string
	if noCache {
		cacheSwitch = "n"
	} else {
		cacheSwitch = "y"
	}
	payload := fmt.Sprintf("%d,%s,%s,%s,%s", auditClient.TransactionNum, auditClient.Command, stockSymbol, userID, cacheSwitch)

	status, body, err := lib.ClientSendRequest(serverurls.Env.QuoteCacheServer, payload)
	if err != nil {
		return 0, errors.New("Failed to get quote: " + err.Error())
	}

	if status != lib.StatusOk {
		return 0, errors.New("Failed to get quote: Response Error: Status " + strconv.Itoa(status) + " " + body)
	}

	// Process Data
	quote, err := strconv.ParseUint(body, 10, 64)
	if err != nil {
		return 0, errors.New("Failed to get quote: " + err.Error())
	}

	return quote, nil
}
