package e2e

import (
	"testing"

	"extremeWorkload.com/daytrader/lib"
)

func setupataValidationTests(t *testing.T) {
	status, body, err := userClient.AddRequest(userid, lib.CentsToDollars(addAmount))
	handleErrors("Add Failed", status, body, err, t)
}

func TestInvalidUserid(t *testing.T) {
	status, body, err := userClient.AddRequest("!!", lib.CentsToDollars(addAmount))
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "Invalid userid" {
		t.Error("Incorrect response returned " + body)
	}
}

func TestUserDoesNotExist(t *testing.T) {
	status, body, err := userClient.BuyRequest("APPLE", "DOG", "5.00")
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "User does not exist" {
		t.Error("Incorrect response returned " + body)
	}
}

func TestZeroAmount(t *testing.T) {
	setupataValidationTests(t)
	status, body, err := userClient.AddRequest(userid, "0.00")
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "Amount cannot be zero" {
		t.Error("Incorrect response returned " + body)
	}
}

func TestNegativeAmount(t *testing.T) {
	setupataValidationTests(t)
	status, body, err := userClient.AddRequest(userid, "-1.00")
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "Invalid amount" {
		t.Error("Incorrect response returned " + body)
	}
}

func TestInvalidAmount(t *testing.T) {
	setupataValidationTests(t)
	status, body, err := userClient.AddRequest(userid, "sdfsd")
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "Invalid amount" {
		t.Error("Incorrect response returned " + body)
	}
}

func TestInvalidStockSymbol(t *testing.T) {
	setupataValidationTests(t)
	status, body, err := userClient.BuyRequest(userid, "ADDS", "1.00")
	checkUserCommandError("Add did not return user error", status, body, err, t)
	if body != "Invalid stock symbol" {
		t.Error("Incorrect response returned " + body)
	}
}
