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
		Server:         "audit-test",
		TransactionNum: 1,
		Command:        "ADD",
	}

	var funds uint64 = 55
	auditClient.LogUserCommandRequest(
		auditclient.UserCommandInfo{
			OptionalUserID:       "testUser",
			OptionalFilename:     "testFile",
			OptionalFundsInCents: &funds,
		},
	)

	auditClient.LogQuoteServerResponse(
		auditclient.QuoteServerResponseInfo{
			PriceInCents:    2,
			StockSymbol:     "ABC",
			UserID:          "testUser",
			QuoteServerTime: 5,
			CryptoKey:       "crypto1",
		},
	)

	auditClient.LogAccountTransaction(
		auditclient.AccountTransactionInfo{
			Action:       "TEST",
			UserID:       "testUser",
			FundsInCents: 5,
		},
	)

	auditClient.LogSystemEvent(
		auditclient.SystemEventInfo{
			OptionalUserID:       "testUser",
			OptionalFilename:     "testFile",
			OptionalFundsInCents: &funds,
		},
	)

	auditClient.LogErrorEvent(
		auditclient.ErrorEventInfo{
			OptionalUserID:       "testUser",
			OptionalFilename:     "testFile",
			OptionalFundsInCents: &funds,
			OptionalErrorMessage: "This is an error event",
		},
	)

	auditClient.LogDebugEvent(
		auditclient.DebugEventInfo{
			OptionalUserID:       "testUser2",
			OptionalFilename:     "testFile",
			OptionalFundsInCents: &funds,
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
