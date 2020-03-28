package main

import (
	"fmt"
	"net"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
)

func failedToReadUserMessage(err error) string {
	return "Failed to read user " + err.Error()
}

func failedToUpdateUserMessage(err error) string {
	return "Failed to update user " + err.Error()
}

func failedToPopStackMessage(stackType string, err error) string {
	return "Failed to pop " + stackType + " stack " + err.Error()
}

func failedToReadTriggerMessage(err error) string {
	return "Failed to read trigger " + err.Error()
}

func failedToCreateTriggerMessage(err error) string {
	return "Failed to create trigger " + err.Error()
}

func failedToUpdateTriggerAmount(err error) string {
	return "Failed to update trigger amount " + err.Error()
}

func failedToUpdateTriggerPrice(err error) string {
	return "Failed to update trigger price " + err.Error()
}

func failedToCancelTrigger(err error) string {
	return "Failed to cancel trigger " + err.Error()
}

func serverSendResponseNoError(conn net.Conn, status int, message string, auditClient *auditclient.AuditClient) {
	err := lib.ServerSendResponse(conn, status, message)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to send response to %s. %d: %s", conn.RemoteAddr().String(), status, message)
		lib.Errorln(errorMessage)

		if auditClient != nil {
			auditClient.LogErrorEvent(errorMessage)
		}
	}
}

func findStockAmount(investments []data.Investment, stockSymbol string) uint64 {
	for _, investment := range investments {
		if investment.Stock == stockSymbol {
			return investment.Amount
		}
	}
	return 0
}
