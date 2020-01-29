package main

import (
	"fmt"
	"testing"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// Note to actually run this test a local/docker
// version of the audit server is required.
func TestAuditServer(t *testing.T) {

	var auditClient = auditclient.AuditClient{
		Server: "audit-test",
	}

	var funds = 55
	auditClient.LogUserCommandRequest(
		auditclient.UserCommandInfo{
			TransactionNum:   1,
			Command:          "TEST",
			OptionalUsername: "testUser",
			OptionalFilename: "testFile",
			OptionalFunds:    &funds,
		},
	)

	auditClient.LogQuoteServerResponse(
		auditclient.QuoteServerResponseInfo{
			TransactionNum:  2,
			Price:           2,
			StockSymbol:     "ABC",
			Username:        "testUser",
			QuoteServerTime: 5,
			CryptoKey:       "crypto1",
		},
	)

	auditClient.LogAccountTransaction(
		auditclient.AccountTransactionInfo{
			TransactionNum: 3,
			Action:         "TEST",
			Username:       "testUser",
			Funds:          5,
		},
	)

	auditClient.LogSystemEvent(
		auditclient.SystemEventInfo{
			TransactionNum:   4,
			Command:          "TEST",
			OptionalUsername: "testUser",
			OptionalFilename: "testFile",
			OptionalFunds:    &funds,
		},
	)

	auditClient.LogErrorEvent(
		auditclient.ErrorEventInfo{
			TransactionNum:       5,
			Command:              "TEST",
			OptionalUsername:     "testUser",
			OptionalFilename:     "testFile",
			OptionalFunds:        &funds,
			OptionalErrorMessage: "This is an error event",
		},
	)

	auditClient.LogDebugEvent(
		auditclient.DebugEventInfo{
			TransactionNum:       6,
			Command:              "TEST",
			OptionalUsername:     "testUser2",
			OptionalFilename:     "testFile",
			OptionalFunds:        &funds,
			OptionalDebugMessage: "This is an debug event",
		},
	)

	fmt.Println("Logs of testUser2")
	logs, _ := auditClient.DumpLog("testUser2")
	fmt.Println(logs)

	fmt.Println("All Logs")
	logs, _ = auditClient.DumpLogAll()
	fmt.Println(logs)
}
