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
		handleAdd(conn, jsonCommand, auditClient)
	case "QUOTE":
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
	case "DISPLAY_SUMMARY":
		handleDisplaySummary(conn, jsonCommand, auditClient)
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid command")
	}

}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	amount := lib.DollarsToCents(jsonCommand.Amount)

	err := dataConn.addAmount(jsonCommand.Userid, amount)
	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "ADD",
		UserID:       jsonCommand.Userid,
		FundsInCents: amount,
	})

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
