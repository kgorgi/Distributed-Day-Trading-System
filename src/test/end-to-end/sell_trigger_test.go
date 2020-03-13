package e2e

import (
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

func setupSellTriggerTest(t *testing.T) {
	status, body, err := userClient.CancelSetSellRequest(userid, stockSymbol)
	checkSystemError("Cancel Sell failed", status, body, err, t)
	summary := getUserSummary(userClient, userid, t)
	if getTestStockTrigger(summary, true) != nil {
		t.Error("Trigger was not cleared initially")
	}
	const amountForSell = (sellAmount / sellTriggerPrice) * quoteValue
	status, body, err = userClient.AddRequest(userid, lib.CentsToDollars(amountForSell))
	handleErrors("Add failed", status, body, err, t)

	status, body, err = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(amountForSell))
	handleErrors("Buy failed", status, body, err, t)

	status, body, err = userClient.CommitBuyRequest(userid)
	handleErrors("Commit buy failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)
	if getTestStockCount(summaryAfter) != getTestStockCount(summary)+(sellAmount/sellTriggerPrice) {
		t.Error("Stocks required for test were not added")
	}

}
func TestTriggerSell(t *testing.T) {

	setupSellTriggerTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)

	status, body, err := userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Set Sell AmountFailed", status, body, err, t)

	status, body, err = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(sellTriggerPrice))
	handleErrors("Set Sell Trigger Failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)

	if getTestStockTrigger(summaryAfter, true) == nil {
		t.Error("Trigger was not saved")
	}

	time.Sleep(65 * time.Second)

	summaryAfter = getUserSummary(userClient, userid, t)

	if getTestStockTrigger(summaryAfter, true) != nil {
		t.Error("Trigger was not cleared")
	}

	expectedStocksSold := (sellAmount / sellTriggerPrice)
	expectedStockCount := getTestStockCount(summaryBefore) - expectedStocksSold
	isEqual(getTestStockCount(summaryAfter), expectedStockCount, "Trigger stock calculation incorrect", t)

	expectedBalance := summaryBefore.Cents + (expectedStocksSold * quoteValue)
	isEqual(summaryAfter.Cents, expectedBalance, "Money was not properly added", t)

	status, body, err = userClient.CancelSetSellRequest(userid, stockSymbol)
	checkSystemError("Cancel Sell failed", status, body, err, t)

}

func TestTriggerSellEditValues(t *testing.T) {
	setupSellTriggerTest(t)
	status, body, err := userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Set Sell Amount failed", status, body, err, t)

	status, body, err = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(sellTriggerPrice))
	handleErrors("Set Sell Trigger failed", status, body, err, t)

	summaryBefore := getUserSummary(userClient, userid, t)
	stocksBefore := getTestStockCount(summaryBefore)

	status, body, err = userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellTriggerPrice-1))
	checkUserCommandError("Should fail when setting amount < trigger price", status, body, err, t)

	status, body, err = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount+1))
	checkUserCommandError("Should fail when setting trigger price > amount", status, body, err, t)

	status, body, err = userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount*2))
	handleErrors("Set Sell Amount failed", status, body, err, t)

	summaryAfter := getUserSummary(userClient, userid, t)
	stocksAfter := getTestStockCount(summaryAfter)

	isEqual(stocksBefore-(sellAmount/sellTriggerPrice), stocksAfter, "Stocks were not properly subtracted from account", t)

	status, body, _ = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(sellTriggerPrice*2))
	handleErrors("Set Sell Trigger failed", status, body, err, t)

	summaryAfter = getUserSummary(userClient, userid, t)
	stocksAfter = getTestStockCount(summaryAfter)

	isEqual(stocksBefore, stocksAfter, "Stocks were not properly added to account", t)

	triggerAfter := getTestStockTrigger(summaryAfter, true)

	isEqual(triggerAfter.Amount_Cents, sellAmount*2, "Amount was not updated", t)
	isEqual(triggerAfter.Price_Cents, sellTriggerPrice*2, "Triggerprice was not updated", t)

	status, body, err = userClient.CancelSetSellRequest(userid, stockSymbol)
	checkSystemError("Cancel Sell failed", status, body, err, t)
}
