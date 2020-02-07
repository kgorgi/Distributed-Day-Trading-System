package e2e

import (
	"strconv"
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

func TestTriggerSell(t *testing.T) {
	userid := "thewolf"
	var addAmount uint64 = 1000234
	var buyAmount uint64 = 500
	var sellAmount uint64 = 500
	var triggerPrice uint64 = 10
	stockSymbol := "DOG"

	status, body, _ := userClient.CancelSetBuyRequest(userid, stockSymbol)
	status, body, _ = userClient.CancelSetSellRequest(userid, stockSymbol)

	status, body, _ = userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	if status != lib.StatusOk {
		t.Error("add failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	if status != lib.StatusOk {
		t.Error("Buy failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.CommitBuyRequest(userid)
	if status != lib.StatusOk {
		t.Error("Commit buy failed\n" + strconv.Itoa(status) + body)
	}

	summaryBefore, err := userClient.GetSummary(userid)
	if err != nil {
		t.Error("Display Summary failed")
	}

	status, body, _ = userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	if status != lib.StatusOk {
		t.Error("Set Sell AmountFailed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(triggerPrice))
	if status != lib.StatusOk {
		t.Error("Set Sell Trigger Failed\n" + strconv.Itoa(status) + body)
	}

	summaryAfter, err := userClient.GetSummary(userid)
	if err != nil {
		t.Error("Display Summary failed")
	}

	if len(summaryAfter.Triggers) != len(summaryBefore.Triggers)+1 {
		t.Error("Trigger was not saved")
	}

	time.Sleep(65 * time.Second)

	summaryAfter, err = userClient.GetSummary(userid)
	if err != nil {
		t.Error("Display Summary failed")
	}

	if len(summaryAfter.Triggers) != len(summaryBefore.Triggers) {
		t.Error("Trigger was not cleared")
	}

	expectedStocksSold := (buyAmount / triggerPrice)
	expectedStockCount := summaryBefore.Investments[0].Amount - expectedStocksSold
	if len(summaryAfter.Investments) > 0 && summaryAfter.Investments[0].Amount != expectedStockCount {
		t.Error("Trigger was not properly executed")
	}

	if summaryAfter.Cents != summaryBefore.Cents+(expectedStocksSold*quoteValue) {
		t.Error("Money from sale was not added to account")
	}

}
