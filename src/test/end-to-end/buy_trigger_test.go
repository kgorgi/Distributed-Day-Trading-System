package e2e

import (
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

func setupBuyTriggerTest(t *testing.T) {
	status, body, err := userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkSystemError("Cancel Buy failed", status, body, err, t)
	summary := getUserSummary(userid, t)
	if getTestStockTrigger(summary, false) != nil {
		t.Error("Trigger was not cleared initially")
	}
	status, body, err = userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add failed", status, body, err, t)
}

func TestTriggerBuy(t *testing.T) {

	setupBuyTriggerTest(t)

	summaryBefore := getUserSummary(userid, t)

	status, body, err := userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Set Buy Amount failed", status, body, err, t)

	status, body, _ = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice))
	handleErrors("Set Buy Trigger failed", status, body, err, t)

	summaryAfter := getUserSummary(userid, t)

	if getTestStockTrigger(summaryAfter, false) == nil {
		t.Error("Trigger was not saved")
	}
	time.Sleep(65 * time.Second)

	summaryAfter = getUserSummary(userid, t)

	if getTestStockTrigger(summaryAfter, false) != nil {
		t.Error("Trigger was not cleared")
	}

	expectedStocksBought := (buyAmount / quoteValue)
	expectedStockCount := getTestStockCount(summaryBefore) + expectedStocksBought
	isEqual(getTestStockCount(summaryAfter), expectedStockCount, "Trigger was not properly executed", t)

	expectedBalance := summaryBefore.Cents - (expectedStocksBought * quoteValue)
	isEqual(summaryAfter.Cents, expectedBalance, "Money was not properly subtracted", t)

}
