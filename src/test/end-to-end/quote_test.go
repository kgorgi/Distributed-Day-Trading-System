package e2e

import (
	"testing"

	"extremeWorkload.com/daytrader/lib"
)

func TestQuote(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	status, body, err = userClient.QuoteRequest(userid, stockSymbol)
	handleErrors("Quote Failed", status, body, err, t)

	if body != "5.00" {
		t.Error("Quote return wrong amount " + body)
	}
}
