package main

import (
	"encoding/json"
	"net"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func processCommand(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	valid := validateParameters(conn, jsonCommand)
	if !valid {
		return
	}

	// Process Command
	switch jsonCommand.Command {
	case "ADD":
		handleAdd(conn, jsonCommand, &auditClient)
	case "QUOTE":
		handleQuote(conn, jsonCommand, &auditClient)
	case "BUY":
		handleBuy(conn, jsonCommand, &auditClient)
	case "COMMIT_BUY":
		handleCommitBuy(conn, jsonCommand, &auditClient)
	case "CANCEL_BUY":
		handleCancelBuy(conn, jsonCommand, &auditClient)
	case "SELL":
		handleSell(conn, jsonCommand, &auditClient)
	case "COMMIT_SELL":
		handleCommitSell(conn, jsonCommand, &auditClient)
	case "CANCEL_SELL":
		handleCancelSell(conn, jsonCommand, &auditClient)
	case "SET_BUY_AMOUNT":
		handleSetBuyAmount(conn, jsonCommand, &auditClient)
	case "SET_BUY_TRIGGER":
		handleSetBuyTrigger(conn, jsonCommand, &auditClient)
	case "CANCEL_SET_BUY":
		handleCancelSetBuy(conn, jsonCommand, &auditClient)
	case "SET_SELL_AMOUNT":
		handleSetSellAmount(conn, jsonCommand, &auditClient)
	case "SET_SELL_TRIGGER":
		handleSetSellTrigger(conn, jsonCommand, &auditClient)
	case "CANCEL_SET_SELL":
		handleCancelSetSell(conn, jsonCommand, &auditClient)
	case "DISPLAY_SUMMARY":
		handleDisplaySummary(conn, jsonCommand, auditClient)
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid command")
	}

}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amount := lib.DollarsToCents(jsonCommand.Amount)

	err := dataConn.addAmount(jsonCommand.Userid, amount, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quote := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}

func handleBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than buy amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &amountInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks

	buyStackMap.push(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks, moneyToRemove)

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy := buyStackMap.pop(jsonCommand.Userid)
	if nextBuy == nil {
		errorMessage := "No Buy to Commit"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	balanceInCents, err := dataConn.getBalance(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if balanceInCents < nextBuy.cents {
		errorMessage := "Account balance is less than stock cost"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	err = dataConn.removeAmount(jsonCommand.Userid, nextBuy.cents, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.addStock(jsonCommand.Userid, nextBuy.stockSymbol, nextBuy.numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO Retry?
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy := buyStackMap.pop(jsonCommand.Userid)
	if nextBuy == nil {
		errorMessage := "No Buy to Cancel"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than sell amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &amountInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToAdd := quoteInCents * numOfStocks

	sellStackMap.push(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks, moneyToAdd)

	lib.ServerSendOKResponse(conn)
}

func handleCommitSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell := sellStackMap.pop(jsonCommand.Userid)
	if nextSell == nil {
		errorMessage := "No Sell to Commit"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	stockAmount, err := dataConn.getStockAmount(jsonCommand.Userid, nextSell.stockSymbol)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if stockAmount < nextSell.numOfStocks {
		errorMessage := "Not enough stock is owned user to sell"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	err = dataConn.removeStock(jsonCommand.Userid, nextSell.stockSymbol, nextSell.numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.addAmount(jsonCommand.Userid, nextSell.cents, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO Return Stock to user's portfolio
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell := sellStackMap.pop(jsonCommand.Userid)
	if nextSell == nil {
		errorMessage := "No Sell to Cancel"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})
		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	balanceInCents, err := dataConn.getBalance(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if amountInCents > balanceInCents {
		errorMessage := "Account balance is less than trigger amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &balanceInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	_, err = dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err == nil {
		errorMessage := "There is an existing trigger, use CANCEL_SET_BUY to update amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &balanceInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	} else if err != ErrDataNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.removeAmount(jsonCommand.Userid, amountInCents, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}
	err = dataConn.createTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, amountInCents, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	triggerPriceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err == ErrDataNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if trigger.Amount_Cents < triggerPriceInCents {
		errorMessage := "Trigger price is too high, no stocks will be able to be bought with current amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	err = dataConn.setTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, triggerPriceInCents, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err == ErrDataNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.addAmount(jsonCommand.Userid, trigger.Amount_Cents, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.deleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO handle error here
		return
	}

	lib.ServerSendOKResponse(conn)

}

func handleSetSellAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if err == nil {
		errorMessage := "There is an existing trigger, use CANCEL_SET_SELL to update amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	} else if err != ErrDataNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	err = dataConn.createTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, amountInCents, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	priceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if err == ErrDataNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if priceInCents > trigger.Amount_Cents {
		errorMessage := "Trigger amount is higher than amount of stocks to sell"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocksOwn, err := dataConn.getStockAmount(jsonCommand.Userid, jsonCommand.StockSymbol)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	numOfStocks := trigger.Amount_Cents / priceInCents

	if numOfStocks > numOfStocksOwn {
		errorMessage := "User does not have enough stocks"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if trigger.Price_Cents != 0 {
		// If there's an existing trigger, give back the reserved stock
		reservedStocks := trigger.Amount_Cents / trigger.Price_Cents
		err := dataConn.addStock(jsonCommand.Userid, jsonCommand.StockSymbol, reservedStocks)
		if err != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
			return
		}
	}

	err = dataConn.removeStock(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.setTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, priceInCents, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO handle stock being removed
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)

	if err == ErrDataNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.deleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO handle error here
		return
	}

	if trigger.Price_Cents == 0 {
		// Trigger not set, no need to re-add stock
		lib.ServerSendOKResponse(conn)
		return
	}

	numOfStocks := trigger.Amount_Cents / trigger.Price_Cents
	err = dataConn.addStock(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleDisplaySummary(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	balanceInCents, err := dataConn.getBalance(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	investments, err := dataConn.getStocks(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	triggers, err := dataConn.getTriggersByUser(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	userDisplay := modelsdata.UserDisplayInfo{}
	userDisplay.Cents = balanceInCents
	userDisplay.Investments = investments

	// convert triggers to triggerdisplayinfos
	triggerDisplays := []modelsdata.TriggerDisplayInfo{}
	for _, trigger := range triggers {
		triggerDisplays = append(
			triggerDisplays,
			modelsdata.TriggerDisplayInfo{trigger.Stock, trigger.Price_Cents, trigger.Amount_Cents, trigger.Is_Sell},
		)
	}
	userDisplay.Triggers = triggerDisplays

	userDisplayBytes, jsonErr := json.Marshal(userDisplay)
	if jsonErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
		return
	}
	lib.ServerSendResponse(conn, lib.StatusOk, string(userDisplayBytes))
}
