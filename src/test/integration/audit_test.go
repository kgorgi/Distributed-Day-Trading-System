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

	auditClient.LogQuoteServerResponse(2, "ABC", "testUser", 5, "crypto1")

	auditClient.LogAccountTransaction("Test", "testUser", 5)

	auditClient.LogSystemEvent()

	auditClient.LogErrorEvent("This is an error event")

	auditClient.LogDebugEvent("This is an debug event")

	fmt.Println("Logs of testUser2")
	logs, _ := auditClient.DumpLog("testUser2")
	fmt.Println(logs)

	fmt.Println("All Logs")
	logs, _ = auditClient.DumpLogAll()
	fmt.Println(logs)
}
