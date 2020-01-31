package main

import (
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
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
		lib.ServerSendResponse(conn, lib.StatusUserError, "Quote price is higher than amount")
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks

	stack := getBuyStack(jsonCommand.Userid)
	reserve := createReserve(jsonCommand.StockSymbol, numOfStocks, moneyToRemove)
	stack.push(reserve, auditClient)

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	stack := getBuyStack(jsonCommand.Userid)
	nextBuy := stack.pop()
	if nextBuy == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "No Buy to Commit")
		return
	}

	balanceInCents, err := dataConn.getBalance(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if balanceInCents < nextBuy.cents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Account balance is less than stock cost")
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
	stack := getBuyStack(jsonCommand.Userid)
	nextBuy := stack.pop()
	if nextBuy == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "No Buy to Cancel")
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Quote price is higher than sell amount")
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToAdd := quoteInCents * numOfStocks

	stack := getSellStack(jsonCommand.Userid)
	reserve := createReserve(jsonCommand.StockSymbol, numOfStocks, moneyToAdd)
	stack.push(reserve, auditClient)

	lib.ServerSendOKResponse(conn)
}

func handleCommitSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	stack := getSellStack(jsonCommand.Userid)
	nextSell := stack.pop()
	if nextSell == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "No Sell to Commit")
		return
	}

	stockAmount, err := dataConn.getStockAmount(jsonCommand.Userid, nextSell.stockSymbol)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if stockAmount < nextSell.numOfStocks {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Not enough stock is owned by the user to sell")
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
	stack := getSellStack(jsonCommand.Userid)
	nextSell := stack.pop()
	if nextSell == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "No Sell to Cancel")
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
		lib.ServerSendResponse(conn, lib.StatusUserError, "Account balance is less than trigger amount")
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
		// TODO replace user's balance
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if trigger == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Trigger amount has not been set")
		return
	}

	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	if trigger.Amount_Cents < amountInCents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Amount too high trigger will never execute")
		return
	}

	err = dataConn.setTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, amountInCents, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if trigger == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Trigger does not exist")
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
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	err := dataConn.createTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, amountInCents, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if trigger == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Trigger amount has not been set")
		return
	}

	priceInCents := lib.DollarsToCents(jsonCommand.Amount)
	if priceInCents > trigger.Amount_Cents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Quote price is higher than amount of stocks to sell")
		return
	}

	numOfStocksOwn, err := dataConn.getStockAmount(jsonCommand.Userid, jsonCommand.StockSymbol)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	numOfStocks := trigger.Amount_Cents / priceInCents

	if numOfStocks > numOfStocksOwn {
		lib.ServerSendResponse(conn, lib.StatusUserError, "User does not have enough stocks")
		return
	}

	err = dataConn.removeStock(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	err = dataConn.setTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, priceInCents, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO handle stock being removed
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, err := dataConn.getTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if trigger == nil {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Trigger does not exist")
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

	stocks, err := dataConn.getStocks(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	var str strings.Builder
	str.WriteString(lib.CentsToDollars(balanceInCents))
	str.WriteString(",")

	for i, element := range stocks {
		str.WriteString(element.stockSymbol)
		str.WriteString(":")
		str.WriteString(strconv.FormatUint(element.numOfStocks, 10))

		if i < len(stocks)-1 {
			str.WriteString(",")
		}
	}

	lib.ServerSendResponse(conn, lib.StatusOk, str.String())
}
