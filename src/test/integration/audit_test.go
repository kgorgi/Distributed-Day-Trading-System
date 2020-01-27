package auditclienttest

import (
	"testing"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// Note to actually run this test a local/docker
// version of the audit server is required.
func TestAuditServerIntegration(t *testing.T) {

	var auditClient auditclient.AuditClient

	var funds = 55
	auditClient.LogUserCommandRequest(
		auditclient.UserCommandInfo{
			Server:         "audit-test",
			TransactionNum: 1,
			Command:        "TEST",
			Username:       "testUser",
			Filename:       "testFile",
			Funds:          &funds,
		},
	)

	auditClient.LogQuoteServerResponse(
		auditclient.QuoteServerResponseInfo{
			Server:          "audit-test",
			TransactionNum:  1,
			Price:           2,
			StockSymbol:     "ABC",
			Username:        "testUser",
			QuoteServerTime: 5,
			CryptoKey:       "crypto1",
		},
	)

	auditClient.LogAccountTransaction(
		auditclient.AccountTransactionInfo{
			Server:         "audit-test",
			TransactionNum: 1,
			Action:         "TEST",
			Username:       "testUser",
			Funds:          5,
		},
	)

	auditClient.LogSystemEvent(
		auditclient.SystemEventInfo{
			Server:         "audit-test",
			TransactionNum: 1,
			Command:        "TEST",
			Username:       "testUser",
			Filename:       "testFile",
			Funds:          &funds,
		},
	)

	auditClient.LogErrorEvent(
		auditclient.ErrorEventInfo{
			Server:         "audit-test",
			TransactionNum: 1,
			Command:        "TEST",
			Username:       "testUser",
			Filename:       "testFile",
			Funds:          &funds,
			ErrorMessage:   "This is an error event",
		},
	)

	auditClient.LogDebugEvent(
		auditclient.DebugEventInfo{
			Server:         "audit-test",
			TransactionNum: 1,
			Command:        "TEST",
			Username:       "testUser",
			Filename:       "testFile",
			Funds:          &funds,
			DebugMessage:   "This is an debug event",
		},
	)
}
