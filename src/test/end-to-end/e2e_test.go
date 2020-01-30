package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"extremeWorkload.com/daytrader/lib"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

func TestBuy(t *testing.T) {
	userid := "thewolf"
	var addAmount uint64 = 1000234
	var buyAmount uint64 = 70000
	stockSymbol := "DOG"

	status, body, _ := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	if status != lib.StatusOk {
		t.Error("add failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.QuoteRequest(userid, stockSymbol)
	if status != lib.StatusOk {
		t.Error("Buy failed\n" + strconv.Itoa(status) + body)
	}
	stockValue := lib.DollarsToCents(body)
	fmt.Println(stockValue)

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	if status != lib.StatusOk {
		t.Error("Buy failed\n" + strconv.Itoa(status) + body)
	}
	fmt.Println(body)

	status, body, _ = userClient.CommitBuyRequest(userid)
	if status != lib.StatusOk {
		t.Error("Commit buy failed\n" + strconv.Itoa(status) + body)
	}

	status, body, _ = userClient.DisplaySummaryRequest(userid)
	if status != lib.StatusOk {
		t.Error("Display summary failed\n" + strconv.Itoa(status) + body)
	}

	stocksBoughtCheck := buyAmount / stockValue
	amountLeftCheck := addAmount - (stocksBoughtCheck * stockValue)

	bodySplit := strings.Split(body, ",")

	if bodySplit[0] != lib.CentsToDollars(amountLeftCheck) {
		t.Error("Total amount remaining in user account is not correct\n" + bodySplit[0] + "!=" + lib.CentsToDollars(amountLeftCheck))
	}

	if bodySplit[1] != stockSymbol+":"+strconv.Itoa(int(stocksBoughtCheck)) {
		t.Error("Stock count does not match\n" + bodySplit[1] + "!=" + stockSymbol + ":" + strconv.Itoa(int(stocksBoughtCheck)))
	}

}
