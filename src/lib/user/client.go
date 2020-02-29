package user

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const webserverAddress = "https://localhost:8080/command/"

const caCert = `-----BEGIN CERTIFICATE-----
MIIDCTCCAfGgAwIBAgIUJEv0GYYf5NgCkSt99YAlApZ/kgQwDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTIwMDMwMTIxMjI0NVoXDTIxMDMw
MTIxMjI0NVowFDESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAvfYpuQ3SD9MqgxmidWSd6SYYqp5sUVtJJBjtMtdg2hp0
lPys3mtMb2/fCWyDzkDew2Ks+TqGk2F4ueHavRZjSbWmJKhoBq1QCLbiIj30OW6O
uYg4a3Ds3B6KS6MmYotkUHgUBYsQ01kK6ofKbMUc3aaLzCf+J6JVa+V6YbX6QVQn
JxAZr2CU18rjIWQofPx0Rt5G/RzyEZx5dQWa7u5JXAvNn7vjOYSjve9xEYkt/jxr
XD0Z5Vucccq0z8rDhj5HAbGRevXQT+S+KRKoiWAE97Brk+coxqfvZ+8a6ZW0jpYy
phfnqaHfEw1uhmRSVzDkZrWSwdtcHN996Q7MpEOVQwIDAQABo1MwUTAdBgNVHQ4E
FgQUQQT2f6J/Sb6T3RP5s/cB3zJQAH8wHwYDVR0jBBgwFoAUQQT2f6J/Sb6T3RP5
s/cB3zJQAH8wDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEACeOn
l9vOsQLnER625OgCPhoLyx8oKwdM+Dh4m2PJCt0vBJnFKCtRrQekd4ryyCSYgg/c
zkQHSwRr4OCA64PiLKfRdNpdYnQ4TrROeL6c06IJ3IlnXtoFeRzYp8T2IlSMwt3L
tFQKsqgie5aMDNtXjr83W9RUUSr/LGFqI3OOihyAWqye0zWSeUoqaFJC2aFBaov9
0na8c3J4UNTFX2S6tlowIRe6RJxROIRWjg/SzdRBpstrImgaHR0DyGCUk7PF/gAO
8An2umJWjNWHHr8aiQ3nlyTImlx6fc8RrA8JRbukD2UcNc4IBfPFYek9vO/li9/B
zHiz81RZYkHq3ixFnQ==
-----END CERTIFICATE-----
`

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

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCert))
	// // Create a HTTPS client and supply the created CA pool and certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
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

func SaveDumplog(body string, filename string) error {

	dumpFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dumpFile.Close()
	_, err = dumpFile.WriteString(body)
	if err != nil {
		return err
	}

	return nil
}

func AddRequest(userid string, amount string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
		Amount: amount,
	}
	return makeRequest("POST", "ADD", createParameters(command))
}

func QuoteRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "QUOTE", createParameters(command))
}

func BuyRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "BUY", createParameters(command))
}

func CommitBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "COMMIT_BUY", createParameters(command))
}

func CancelBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "CANCEL_BUY", createParameters(command))
}

func SellRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SELL", createParameters(command))
}

func CommitSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "COMMIT_SELL", createParameters(command))
}

func CancelSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "CANCEL_SELL", createParameters(command))
}

func SetBuyAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_BUY_AMOUNT", createParameters(command))
}

func CancelSetBuyRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "CANCEL_SET_BUY", createParameters(command))
}

func SetBuyTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_BUY_TRIGGER", createParameters(command))
}

func SetSellAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_SELL_AMOUNT", createParameters(command))
}

func CancelSetSellRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return makeRequest("GET", "CANCEL_SET_SELL", createParameters(command))
}

func SetSellTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return makeRequest("POST", "SET_SELL_TRIGGER", createParameters(command))
}

func DumplogRequest(userid string, filename string) (int, string, error) {
	var command = commandParams{
		UserID:   userid,
		Filename: filename,
	}
	return makeRequest("POST", "DUMPLOG", createParameters(command))
}

func DisplaySummaryRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return makeRequest("POST", "DISPLAY_SUMMARY", createParameters(command))
}
