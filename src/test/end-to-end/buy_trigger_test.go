package e2e

import (
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

func setupBuyTriggerTest(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add failed", status, body, err, t)

	status, body, err = userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkSystemError("Cancel Buy failed", status, body, err, t)
	summary := getUserSummary(userClient, userid, t)
	if getTestStockTrigger(summary, false) != nil {
		t.Error("Trigger was not cleared initially")
	}
}

func TestTriggerBuyNotEnough(t *testing.T) {
	setupBuyTriggerTest(t)
	summaryBefore := getUserSummary(userClient, userid, t)

	status, body, err := userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(summaryBefore.Cents+10))
	checkUserCommandError("Should have failed for not having enought amount", status, body, err, t)

}

func TestTriggerBuy(t *testing.T) {

	setupBuyTriggerTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)

	status, body, err := userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice))
	checkUserCommandError("Should have failed for not having amoutn set", status, body, err, t)

	// + 1 is to make sure leftover reserve money is returned
	status, body, err = userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount+1))
	handleErrors("Set Buy Amount failed", status, body, err, t)

	status, body, err = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice))
	handleErrors("Set Buy Trigger failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)

	if getTestStockTrigger(summaryAfter, false) == nil {
		t.Error("Trigger was not saved")
	}
	time.Sleep(65 * time.Second)

	summaryAfter = getUserSummary(userClient, userid, t)

	if getTestStockTrigger(summaryAfter, false) != nil {
		t.Error("Trigger was not cleared")
	}

	expectedStocksBought := (buyAmount / quoteValue)
	expectedStockCount := getTestStockCount(summaryBefore) + expectedStocksBought
	isEqual(getTestStockCount(summaryAfter), expectedStockCount, "Trigger was not properly executed", t)

	expectedBalance := summaryBefore.Cents - (expectedStocksBought * quoteValue)
	isEqual(summaryAfter.Cents, expectedBalance, "Money was not properly subtracted", t)

	status, body, err = userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkSystemError("Cancel Buy failed", status, body, err, t)
}

func TestTriggerBuyEditValues(t *testing.T) {
	setupBuyTriggerTest(t)
	status, body, err := userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Set Buy Amount failed", status, body, err, t)

	status, body, _ = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice))
	handleErrors("Set Buy Trigger failed", status, body, err, t)

	summaryBefore := getUserSummary(userClient, userid, t)

	status, body, err = userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice-1))
	checkUserCommandError("Should fail when setting amount < trigger price", status, body, err, t)

	status, body, err = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount+1))
	checkUserCommandError("Should fail when setting trigger price > amount", status, body, err, t)

	status, body, err = userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount*2))
	handleErrors("Set Buy Amount failed", status, body, err, t)

	status, body, _ = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice*2))
	handleErrors("Set Buy Trigger failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)

	isEqual(summaryBefore.Cents-buyAmount, summaryAfter.Cents, "Money was not subtracted from account", t)

	triggerAfter := getTestStockTrigger(summaryAfter, false)

	isEqual(triggerAfter.Amount_Cents, buyAmount*2, "Amount was not updated", t)
	isEqual(triggerAfter.Price_Cents, buyTriggerPrice*2, "Triggerprice was not updated", t)

	status, body, err = userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkSystemError("Cancel Buy failed", status, body, err, t)
}

func TestTriggerBuyCancel(t *testing.T) {
	setupBuyTriggerTest(t)

	status, body, err := userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkUserCommandError("Expected a user error response", status, body, err, t)

	summaryBefore := getUserSummary(userClient, userid, t)

	status, body, err = userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Set Buy Amount failed", status, body, err, t)

	status, body, _ = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(buyTriggerPrice))
	handleErrors("Set Buy Trigger failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)
	isEqual(summaryAfter.Cents, summaryBefore.Cents-buyAmount, "Money was not witheld", t)

	status, body, err = userClient.CancelSetBuyRequest(userid, stockSymbol)
	handleErrors("Cancel Buy failed", status, body, err, t)

	summaryAfter = getUserSummary(userClient, userid, t)
	isEqual(summaryAfter.Cents, summaryBefore.Cents, "Money was not returned", t)

	status, body, err = userClient.CancelSetBuyRequest(userid, stockSymbol)
	checkSystemError("Cancel Buy failed", status, body, err, t)
}
