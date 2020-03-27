package main

import (
	"errors"
	"regexp"
	"strconv"
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

func validateParameters(commandMap map[string]string) error {
	if commandMap["command"] == "DUMPLOG" && commandMap["userid"] == "" {
		return nil
	}

	// Check userID has valid characters
	if !isAlphanumeric(commandMap["userid"]) {
		return errors.New("Invalid userid")
	}

	// Validate StockSymbol
	if _, ok := noStockSymbolParameter[commandMap["command"]]; !ok {
		if !isStockSymbol(commandMap["stockSymbol"]) {
			return errors.New("Invalid stockSymbol")
		}
	}

	// Validate Amount
	if _, ok := noAmountParameter[commandMap["command"]]; !ok {
		if !isAmount(commandMap["amount"]) {
			return errors.New("Invalid amount")
		}

		amount, err := strconv.ParseFloat(commandMap["amount"], 64)
		if err != nil {
			return errors.New("Invalid amount")
		}

		if amount == 0 {
			return errors.New("Amount cannot be zero")
		}

		if amount <= 0 {
			return errors.New("Amount cannot be less than zero")
		}
	}

	return nil
}
