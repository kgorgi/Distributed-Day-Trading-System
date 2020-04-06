package e2e

import (
	"fmt"
	"testing"

	"extremeWorkload.com/daytrader/lib"
)

func TestAddNewUser(t *testing.T) {
	userid := "user12"
	status, body, err := userClient.DisplaySummaryRequest(userid)
	if err != nil {
		t.Error("Failed display summary request " + err.Error())
	}

	if status != lib.StatusUserError {
		fmt.Println("User already exists, skipping add new user test")
		return
	}

	status, body, err = userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summary := getUserSummary(userClient, userid, t)
	isEqual(summary.Cents, addAmount, "Incorrect amount added to new user", t)
	isEqual(uint64(len(summary.Investments)), 0, "User has investments", t)
}

func TestAddAmountToUser(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)

	summaryBefore := getUserSummary(userClient, userid, t)
	status, body, err = userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	summaryAfter := getUserSummary(userClient, userid, t)

	isEqual(summaryAfter.Cents, summaryBefore.Cents+addAmount, "Incorrect amount added to user", t)
}
