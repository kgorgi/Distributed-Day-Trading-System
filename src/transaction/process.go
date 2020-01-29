package main

import (
	"net"

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
		handleAdd(conn, jsonCommand, auditClient)
	case "QUOTE":
		handleQuote(conn, jsonCommand, auditClient)
	case "BUY":
	case "COMMIT_BUY":
	case "CANCEL_BUY":
	case "SELL":
	case "COMMIT_SELL":
	case "CANCEL_SELL":
	case "SET_BUY_AMOUNT":
	case "CANCEL_SET_BUY":
	case "SET_BUY_TRIGGER":
	case "SET_SELL_AMOUNT":
	case "CANCEL_SET_SELL":
	case "SET_SELL_TRIGGER":
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid command")
	}

}

func getCents(dollarString string) uint64 {
	return 1
}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	amount := getCents(jsonCommand.Amount)

	err := dataConn.addAmount(jsonCommand.Userid, amount)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	// auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
	// 	Action:       "ADD",
	// 	UserID:       jsonCommand.Userid,
	// 	FundsInCents: amount,
	// })

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	quote := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}
