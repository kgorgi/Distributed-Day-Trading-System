package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	user "extremeWorkload.com/daytrader/lib/user"
)

var userClient *user.UserClient

const webserverAddress = "https://localhost:8080/"

func TestMain(m *testing.M) {
	var err error
	userClient, err = user.CreateClient(webserverAddress, os.Getenv("CLIENT_SSL_CERT_LOCATION"))
	if err != nil {
		fmt.Println("Failed while creating a user client")
		fmt.Println(err)
		os.Exit(1)
		return
	}
	os.Exit(m.Run())
}

func setupBuyTest(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)
}

func TestBuyDoesNotModifyAccount(t *testing.T) {
	setupBuyTest(t)
	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	summaryAfterBuy := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterBuy)
	isEqual(summaryAfterBuy.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)

	status, body, err = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)
}

func TestBuyWithCommit(t *testing.T) {
	setupBuyTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	status, body, err = userClient.CommitBuyRequest(userid)
	handleErrors("Commit Buy Failed", status, body, err, t)

	summaryAfterCommit := getUserSummary(userClient, userid, t)

	stocksBoughtCheck := buyAmount / quoteValue
	amountLeftCheck := existingBalance - (stocksBoughtCheck * quoteValue)
	newStock := getTestStockCount(summaryAfterCommit)

	isEqual(summaryAfterCommit.Cents, amountLeftCheck, "Total amount remaining in user account is not correct", t)
	isEqual(newStock, stocksBoughtCheck+existingStock, "Stock count does not match", t)
}

func TestBuyWithCancel(t *testing.T) {
	setupBuyTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	status, body, err = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)

	summaryAfterCancel := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterCancel)

	isEqual(summaryAfterCancel.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)
}

func TestBuyTimeout(t *testing.T) {
	setupBuyTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	time.Sleep(61 * time.Second)

	summaryAfterTimeout := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterTimeout)

	isEqual(summaryAfterTimeout.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)

	status, body, err = userClient.CommitBuyRequest(userid)
	checkUserCommandError("Commit Buy Succeeded", status, body, err, t)

	status, body, err = userClient.CancelBuyRequest(userid)
	checkUserCommandError("Cancel Buy Succeeded", status, body, err, t)
}

func TestCommitBuyFailsWithoutBuy(t *testing.T) {
	status, body, err := userClient.CommitBuyRequest(userid)
	checkUserCommandError("Commit Buy Succeeded", status, body, err, t)
}

func TestCancelBuyFailsWithoutBuy(t *testing.T) {
	status, body, err := userClient.CancelBuyRequest(userid)
	checkUserCommandError("Cancel Buy Succeeded", status, body, err, t)
}

func TestBuyStack(t *testing.T) {
	setupBuyTest(t)

	// Add to Stack
	for i := 0; i < 4; i++ {
		status, body, err := userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
		handleErrors("Buy Failed", status, body, err, t)
	}

	// Pop from stack
	status, body, err := userClient.CommitBuyRequest(userid)
	handleErrors("Commit Buy Failed", status, body, err, t)

	status, body, err = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)

	status, body, err = userClient.CommitBuyRequest(userid)
	handleErrors("Commit Buy Failed", status, body, err, t)

	status, body, err = userClient.CancelBuyRequest(userid)
	handleErrors("Cancel Buy Failed", status, body, err, t)

	// Extra Commands which should fail
	status, body, err = userClient.CommitBuyRequest(userid)
	checkUserCommandError("Commit Buy Succeeded", status, body, err, t)

	status, body, err = userClient.CommitBuyRequest(userid)
	checkUserCommandError("Commit Buy Succeeded", status, body, err, t)
}
