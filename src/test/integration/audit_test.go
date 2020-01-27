package auditclienttest

import (
	"fmt"
	"testing"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// Note to actually run this test a local/docker
// version of the audit server is required.
func TestAuditServerIntegration(t *testing.T) {

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
			Username:             "testUser",
			OptionalFilename:     "testFile",
			OptionalFunds:        &funds,
			OptionalDebugMessage: "This is an debug event",
		},
	)

	logs, _ := auditClient.DumpLogAll()

	fmt.Println(logs)
}
