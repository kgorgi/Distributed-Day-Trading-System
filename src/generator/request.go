package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type commandParams struct {
	UserID      string
	Amount      string
	StockSymbol string
	Filename    string
}

func createParameters(command commandParams) url.Values {
	params := make(url.Values)
	params.Add("userid", command.UserID)

	if len(command.Amount) > 0 {
		params.Add("amount", command.Amount)
	}
	if len(command.StockSymbol) > 0 {
		params.Add("stocksymbol", command.StockSymbol)
	}
	if len(command.Filename) > 0 {
		params.Add("filename", command.StockSymbol)
	}
	return params
}

func makeRequest(httpMethod string, command string, params url.Values) (int, string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(httpMethod, webserverAddress+command+"?"+params.Encode(), nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	return resp.StatusCode, string(body), nil
}

func addRequest(userid string, amount string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
		Amount: amount,
	}
	return makeRequest("POST", "ADD", createParameters(command))
}

func quoteRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "QUOTE", createParameters(command))
}

func buyRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "BUY", createParameters(command))
}

func commitBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "COMMIT_BUY", createParameters(command))
}

func cancelBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "CANCEL_BUY", createParameters(command))
}

func sellRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SELL", createParameters(command))
}

func commitSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "COMMIT_SELL", createParameters(command))
}

func cancelSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "CANCEL_SELL", createParameters(command))
}

func setBuyAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_BUY_AMOUNT", createParameters(command))
}

func cancelSetBuyRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "CANCEL_SET_BUY", createParameters(command))
}

func setBuyTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_BUY_TRIGGER", createParameters(command))
}

func setSellAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_SELL_AMOUNT", createParameters(command))
}

func cancelSetSellRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "CANCEL_SET_SELL", createParameters(command))
}

func setSellTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_SELL_TRIGGER", createParameters(command))
}

func dumplogRequest(userid string, filename string) (int, string, error) {
	var command = commandParams{
		UserID:   userid,
		Filename: filename,
	}
	return makeRequest("POST", "DUMPLOG", createParameters(command))
}

func displaySummaryRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "DISPLAY_SUMMARY", createParameters(command))
}
