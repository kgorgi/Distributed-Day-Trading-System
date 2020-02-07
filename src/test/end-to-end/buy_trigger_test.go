package main

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

func TestTriggerBuy(t *testing.T) {
	userid := "thewolf"
	var addAmount uint64 = 1000234
	var buyAmount uint64 = 5000
	var triggerPrice uint64 = 500
	stockSymbol := "DOG"

	status, body, _ := userClient.CancelSetBuyRequest(userid, stockSymbol)
	status, body, _ = userClient.CancelSetSellRequest(userid, stockSymbol)

	status, body, _ = userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	if status != lib.StatusOk {
		t.Error("add failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	var summaryBefore modelsdata.UserDisplayInfo
	json.Unmarshal([]byte(body), &summaryBefore)

	status, body, _ = userClient.SetBuyAmountRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	if status != lib.StatusOk {
		t.Error("Set Buy Amount failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.SetBuyTriggerRequest(userid, stockSymbol, lib.CentsToDollars(triggerPrice))
	if status != lib.StatusOk {
		t.Error("Set Buy Trigger failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	var summaryAfter modelsdata.UserDisplayInfo
	json.Unmarshal([]byte(body), &summaryAfter)

	if len(summaryAfter.Triggers) != len(summaryBefore.Triggers)+1 {
		t.Error("Trigger was not saved")
	}

	time.Sleep(65 * time.Second)

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	json.Unmarshal([]byte(body), &summaryAfter)

	if len(summaryAfter.Triggers) != len(summaryBefore.Triggers) {
		t.Error("Trigger was not cleared")
	}

	expectedStocksBought := (buyAmount / quoteValue)
	expectedStockCount := summaryBefore.Investments[0].Amount + expectedStocksBought
	if len(summaryAfter.Investments) > 0 && summaryAfter.Investments[0].Amount != expectedStockCount {
		t.Error("Trigger was not properly executed")
	}

	expectedBalance := summaryBefore.Cents - (expectedStocksBought * quoteValue)
	if summaryAfter.Cents != expectedBalance {
		t.Error("Money was not properly subtracted")
	}

}

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

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}
	var summaryBefore modelsdata.UserDisplayInfo
	json.Unmarshal([]byte(body), &summaryBefore)

	status, body, _ = userClient.SetSellAmountRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	if status != lib.StatusOk {
		t.Error("Set Sell AmountFailed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.SetSellTriggerRequest(userid, stockSymbol, lib.CentsToDollars(triggerPrice))
	if status != lib.StatusOk {
		t.Error("Set Sell Trigger Failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	var summaryAfter modelsdata.UserDisplayInfo
	json.Unmarshal([]byte(body), &summaryAfter)

	if len(summaryAfter.Triggers) != len(summaryBefore.Triggers)+1 {
		t.Error("Trigger was not saved")
	}

	time.Sleep(65 * time.Second)

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	json.Unmarshal([]byte(body), &summaryAfter)

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
