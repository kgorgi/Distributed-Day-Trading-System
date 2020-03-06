package user

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"fmt"
)

const webserverAddress = "https://localhost:9090/command/"

var caCertPool *x509.CertPool

func initCertPool() error {
	envCaCertLocation := os.Getenv("CLIENT_SSL_CERT_LOCATION")
	fmt.Println(envCaCertLocation)
	caCert, err := ioutil.ReadFile(envCaCertLocation)
	if err != nil {
		return err
	}

	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return nil
}

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
		params.Add("stockSymbol", command.StockSymbol)
	}
	if len(command.Filename) > 0 {
		params.Add("filename", command.StockSymbol)
	}
	return params
}

func makeRequest(httpMethod string, command string, params url.Values) (int, string, error) {
	if caCertPool == nil{
		initCertPool()
	}
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
