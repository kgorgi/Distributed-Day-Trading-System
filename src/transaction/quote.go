package main

import (
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// GetQuote returns a quote from the quote server
func GetQuote(
	stockSymbol string,
	userID string,
	auditClient auditclient.AuditClient) uint64 {

	return 5
}
