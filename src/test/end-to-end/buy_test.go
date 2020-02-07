package e2e

import (
	"strconv"
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

func TestBuyDoesNotModifyAccount(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summaryBefore := getUserSummary(userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	summaryAfterBuy := getUserSummary(userid, t)
	newStock := getTestStockCount(summaryAfterBuy)
	isEqual(summaryAfterBuy.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)

	status, body, err = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)
}

func TestBuyWithCommit(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summaryBefore := getUserSummary(userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	status, body, _ = userClient.CommitBuyRequest(userid)
	handleErrors("Commit Buy Failed", status, body, err, t)

	summaryAfterCommit := getUserSummary(userid, t)

	stocksBoughtCheck := buyAmount / quoteValue
	amountLeftCheck := existingBalance - (stocksBoughtCheck * quoteValue)
	newStock := getTestStockCount(summaryAfterCommit)

	isEqual(summaryAfterCommit.Cents, amountLeftCheck, "Total amount remaining in user account is not correct", t)
	isEqual(newStock, stocksBoughtCheck+existingStock, "Stock count does not match", t)
}

func TestBuyWithCancel(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summaryBefore := getUserSummary(userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	status, body, _ = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)

	summaryAfterCancel := getUserSummary(userid, t)
	newStock := getTestStockCount(summaryAfterCancel)

	isEqual(summaryAfterCancel.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)
}

func TestBuyTimeout(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summaryBefore := getUserSummary(userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	time.Sleep(61 * time.Second)

	summaryAfterTimeout := getUserSummary(userid, t)
	newStock := getTestStockCount(summaryAfterTimeout)

	isEqual(summaryAfterTimeout.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)
}

func TestBuyCommitFailsWithoutBuy(t *testing.T) {
	status, body, err := userClient.CommitBuyRequest(userid)
	if err != nil {
		t.Error("Commit Buy Failed" + "\n" + err.Error())
	}

	if status != lib.StatusUserError {
		t.Error("Invalid error code returned\n" + strconv.Itoa(status) + " " + body)
	}
}

func TestBuyCancelFailsWithoutBuy(t *testing.T) {
	status, body, err := userClient.CancelBuyRequest(userid)
	if err != nil {
		t.Error("Cancel Buy Failed" + "\n" + err.Error())
	}

	if status != lib.StatusUserError {
		t.Error("Invalid error code returned\n" + strconv.Itoa(status) + " " + body)
	}
}
