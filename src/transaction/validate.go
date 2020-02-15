package main

import (
	"net"
	"regexp"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

var noStockSymbolParameter = make(map[string]bool)
var noAmountParameter = make(map[string]bool)

func initParameterMaps() {
	noStockSymbolParameter["ADD"] = true
	noStockSymbolParameter["COMMIT_BUY"] = true
	noStockSymbolParameter["CANCEL_BUY"] = true
	noStockSymbolParameter["COMMIT_SELL"] = true
	noStockSymbolParameter["CANCEL_SELL"] = true
	noStockSymbolParameter["DISPLAY_SUMMARY"] = true

	noAmountParameter["QUOTE"] = true
	noAmountParameter["COMMIT_BUY"] = true
	noAmountParameter["CANCEL_BUY"] = true
	noAmountParameter["COMMIT_SELL"] = true
	noAmountParameter["CANCEL_SELL"] = true
	noAmountParameter["CANCEL_SELL"] = true
	noAmountParameter["CANCEL_SET_BUY"] = true
	noAmountParameter["CANCEL_SET_SELL"] = true
	noAmountParameter["DISPLAY_SUMMARY"] = true
}

var isAlphanumeric = regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
var isStockSymbol = regexp.MustCompile(`^[A-Z][A-Z]?[A-Z]?$`).MatchString
var isAmount = regexp.MustCompile(`^[0-9]+\.[0-9][0-9]$`).MatchString

func validateParameters(conn net.Conn, commandJSON CommandJSON) bool {
	// Check userID has valid characters
	if !isAlphanumeric(commandJSON.Userid) {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid userid")
		return false
	}

	// Validate user exists
	exists := true
	_, readErr := dataClient.ReadUser(commandJSON.Userid)
	if readErr == dataclient.ErrNotFound {
		exists = false
	} else if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return false
	}

	if commandJSON.Command != "ADD" && !exists {
		lib.ServerSendResponse(conn, lib.StatusUserError, "User does not exist")
		return false
	} else if commandJSON.Command == "ADD" && !exists {
		createErr := dataClient.CreateUser(modelsdata.User{commandJSON.Userid, 0, []modelsdata.Investment{}})
		if createErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
			return false
		}
	}

	// Validate StockSymbol
	if _, ok := noStockSymbolParameter[commandJSON.Command]; !ok {
		if !isStockSymbol(commandJSON.StockSymbol) {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid stockSymbol")
			return false
		}
	}

	// Validate Amount
	if _, ok := noAmountParameter[commandJSON.Command]; !ok {
		if !isAmount(commandJSON.Amount) {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid amount")
			return false
		}

		amount, err := strconv.ParseFloat(commandJSON.Amount, 64)
		if err != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid amount")
			return false
		}

		if amount == 0 {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Amount cannot be zero")
			return false
		}

		if amount <= 0 {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Amount cannot be less than zero")
			return false
		}
	}

	return true
}
