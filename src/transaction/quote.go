package main

import (
	"errors"
	"fmt"
	"strconv"

	"extremeWorkload.com/daytrader/lib/quote"

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
		errorMessage := fmt.Sprintf("Failed to contact quote cache server %s. Bypassing...", err.Error())
		auditClient.LogErrorEvent(errorMessage)
		cents, _, err := quote.Request(stockSymbol, userID, auditClient)
		return cents, err
	}

	if status != lib.StatusOk {
		errorMessage := fmt.Sprintf("Quote cache server did not return OK: %d. Error: %s Bypassing...", status, body)
		auditClient.LogErrorEvent(errorMessage)
		cents, _, err := quote.Request(stockSymbol, userID, auditClient)
		return cents, err
	}

	// Process Data
	quote, err := strconv.ParseUint(body, 10, 64)
	if err != nil {
		return 0, errors.New("Failed to get quote: " + err.Error())
	}

	return quote, nil
}
