package e2e

import (
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

func setupSellTest(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	status, body, _ = userClient.BuyRequest(userid, stockSymbol, lib.CentsToDollars(buyAmount))
	handleErrors("Buy Failed", status, body, err, t)

	status, body, _ = userClient.CommitBuyRequest(userid)
	handleErrors("Commit Buy Failed", status, body, err, t)
}

func TestSellDoesNotModifyAccount(t *testing.T) {
	setupSellTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.SellRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Sell Failed", status, body, err, t)

	summaryAfterSell := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterSell)
	isEqual(summaryAfterSell.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)

	status, body, err = userClient.CancelSellRequest(userid)
	handleErrors("Cancel Sell Failed", status, body, err, t)
}

func TestSellWithCommit(t *testing.T) {
	setupSellTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStocks := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.SellRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Sell Failed", status, body, err, t)

	status, body, err = userClient.CommitSellRequest(userid)
	handleErrors("Commit Sell Failed", status, body, err, t)

	summaryAfterCommit := getUserSummary(userClient, userid, t)

	stocksSold := sellAmount / quoteValue
	amountStocksLeft := existingStocks - stocksSold

	newAmount := stocksSold*quoteValue + existingBalance
	newStock := getTestStockCount(summaryAfterCommit)

	isEqual(summaryAfterCommit.Cents, newAmount, "Total amount remaining in user account is not correct", t)
	isEqual(newStock, amountStocksLeft, "Stock count does not match", t)
}

func TestSellWithCancel(t *testing.T) {
	setupSellTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.SellRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Sell Failed", status, body, err, t)

	status, body, err = userClient.CancelSellRequest(userid)
	handleErrors("Cancel Sell Failed", status, body, err, t)

	summaryAfterCancel := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterCancel)

	isEqual(summaryAfterCancel.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)
}

func TestSellTimeout(t *testing.T) {
	setupSellTest(t)

	summaryBefore := getUserSummary(userClient, userid, t)
	existingStock := getTestStockCount(summaryBefore)
	existingBalance := summaryBefore.Cents

	status, body, err := userClient.SellRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
	handleErrors("Sell Failed", status, body, err, t)

	time.Sleep(61 * time.Second)

	summaryAfterTimeout := getUserSummary(userClient, userid, t)
	newStock := getTestStockCount(summaryAfterTimeout)

	isEqual(summaryAfterTimeout.Cents, existingBalance, "User's account balance was modified", t)
	isEqual(newStock, existingStock, "User's portfolio was modified", t)

	status, body, err = userClient.CommitSellRequest(userid)
	checkUserCommandError("Commit Sell Succeeded", status, body, err, t)

	status, body, err = userClient.CancelSellRequest(userid)
	checkUserCommandError("Cancel Sell Succeeded", status, body, err, t)
}

func TestCommitSellFailsWithoutSell(t *testing.T) {
	status, body, err := userClient.CommitSellRequest(userid)
	checkUserCommandError("Commit Sell Succeeded", status, body, err, t)
}

func TestCancelSellFailsWithoutSell(t *testing.T) {
	status, body, err := userClient.CancelSellRequest(userid)
	checkUserCommandError("Cancel Sell Succeeded", status, body, err, t)
}

func TestSellStack(t *testing.T) {
	setupSellTest(t)

	// Add to Stack
	for i := 0; i < 4; i++ {
		status, body, err := userClient.SellRequest(userid, stockSymbol, lib.CentsToDollars(sellAmount))
		handleErrors("Sell Failed", status, body, err, t)
	}

	// Pop from stack
	status, body, err := userClient.CommitSellRequest(userid)
	handleErrors("Commit Sell Failed", status, body, err, t)

	status, body, err = userClient.CancelSellRequest(userid)
	handleErrors("Cancel Sell Failed", status, body, err, t)

	status, body, err = userClient.CommitSellRequest(userid)
	handleErrors("Commit Sell Failed", status, body, err, t)

	status, body, err = userClient.CancelSellRequest(userid)
	handleErrors("Cancel Sell Failed", status, body, err, t)

	// Extra Commands which should fail
	status, body, err = userClient.CommitSellRequest(userid)
	checkUserCommandError("Commit Sell Succeeded", status, body, err, t)

	status, body, err = userClient.CommitSellRequest(userid)
	checkUserCommandError("Commit Sell Succeeded", status, body, err, t)
}
