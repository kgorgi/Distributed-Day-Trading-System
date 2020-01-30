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
	case "CANCEL_SET_BUY":
	case "SET_BUY_TRIGGER":
	case "SET_SELL_AMOUNT":
	case "CANCEL_SET_SELL":
	case "SET_SELL_TRIGGER":
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

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       jsonCommand.Userid,
		FundsInCents: amount,
	})

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quote := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}

func handleBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	balanceInCents, err := dataConn.getBalance(jsonCommand.Userid)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	if balanceInCents < amountInCents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Balance is less than amount")
		return
	}

	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		lib.ServerSendResponse(conn, lib.StatusUserError, "Quote price is higher than amount")
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks
	err = dataConn.removeAmount(jsonCommand.Userid, moneyToRemove, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	stack := getBuyStack(jsonCommand.Userid)
	reserve := createReseve(jsonCommand.StockSymbol, numOfStocks, moneyToRemove)
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

	err := dataConn.addStock(jsonCommand.Userid, nextBuy.stockSymbol, nextBuy.numOfStocks)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		// TODO Return Amount to users balance
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

	err := dataConn.addAmount(jsonCommand.Userid, nextBuy.cents, auditClient)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
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
	reserve := createReseve(jsonCommand.StockSymbol, numOfStocks, moneyToAdd)
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
