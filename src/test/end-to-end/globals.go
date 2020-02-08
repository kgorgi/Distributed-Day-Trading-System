package e2e

import (
	"encoding/json"
	"strconv"
	"testing"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	userClient "extremeWorkload.com/daytrader/lib/user"
)

const userid = "thewolf"
const stockSymbol = "DOG"
const quoteValue = 500
const addAmount = uint64(1000234)
const buyAmount = uint64(70000)
const sellAmount = uint64(1000)

func getUserSummary(userid string, t *testing.T) modelsdata.UserDisplayInfo {
	status, body, err := userClient.DisplaySummaryRequest(userid)
	handleErrors("Display summary failed", status, body, err, t)

	var summary modelsdata.UserDisplayInfo
	err = json.Unmarshal([]byte(body), &summary)

	if err != nil {
		t.Error("Display summary failed to unmarshal JSON: " + err.Error())
	}
	return summary
}

func handleErrors(errMessage string, status int, body string, err error, t *testing.T) {
	if err != nil {
		t.Error(errMessage + "\n" + err.Error())
	}

	if status != lib.StatusOk {
		t.Error(errMessage + "\n" + strconv.Itoa(status) + " " + body)
	}
}

func checkUserCommandError(errMessage string, status int, body string, err error, t *testing.T) {
	if err != nil {
		t.Error(errMessage + "\n" + err.Error())
	}

	if status != lib.StatusUserError {
		t.Error(errMessage + "\n" + strconv.Itoa(status) + " " + body)
	}
}

func isEqual(a uint64, b uint64, errMessage string, t *testing.T) {
	if a != b {
		t.Error(errMessage + "\n " +
			strconv.FormatUint(a, 10) + "!=" + strconv.FormatUint(b, 10))
	}
}

func getTestStockCount(summary modelsdata.UserDisplayInfo) uint64 {
	var stockNum = uint64(0)
	for _, investment := range summary.Investments {
		if investment.Stock == stockSymbol {
			stockNum = investment.Amount
			break
		}
	}

	return stockNum
}
