package user

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type UserClient struct {
	WebServerAddress string
	Client           *http.Client
}

var caCert []byte

func CreateClient(webServerAddress string, envCaCertLocation string) (*UserClient, error) {
	if len(caCert) == 0 {
		var err error
		caCert, err = ioutil.ReadFile(envCaCertLocation)
		if err != nil {
			return nil, err
		}
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
			ForceAttemptHTTP2: true,
		},
		Timeout: 120 * time.Second,
	}

	return &UserClient{webServerAddress, client}, nil
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

func (client *UserClient) makeRequest(httpMethod string, command string, params url.Values) (int, string, error) {
	req, err := http.NewRequest(httpMethod, client.WebServerAddress+"command/"+command+"?"+params.Encode(), nil)
	if err != nil {
		return 0, "", err
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return 0, "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	resp.Body.Close()

	return resp.StatusCode, string(body), nil
}

func (client *UserClient) HeartRequest() (int, string, error) {
	resp, err := client.Client.Get(client.WebServerAddress + "heartbeat")
	if err != nil {
		return 0, "", err
	}

	resp.Body.Close()
	return resp.StatusCode, "", nil
}

func SaveDumplog(body string, filename string) error {
	dumpFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = dumpFile.WriteString(body)
	if err != nil {
		return err
	}

	dumpFile.Close()

	return nil
}

func (client *UserClient) AddRequest(userid string, amount string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
		Amount: amount,
	}
	return client.makeRequest("POST", "ADD", createParameters(command))
}

func (client *UserClient) QuoteRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("GET", "QUOTE", createParameters(command))
}

func (client *UserClient) BuyRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "BUY", createParameters(command))
}

func (client *UserClient) CommitBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return client.makeRequest("POST", "COMMIT_BUY", createParameters(command))
}

func (client *UserClient) CancelBuyRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return client.makeRequest("POST", "CANCEL_BUY", createParameters(command))
}

func (client *UserClient) SellRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "SELL", createParameters(command))
}

func (client *UserClient) CommitSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return client.makeRequest("POST", "COMMIT_SELL", createParameters(command))
}

func (client *UserClient) CancelSellRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return client.makeRequest("POST", "CANCEL_SELL", createParameters(command))
}

func (client *UserClient) SetBuyAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "SET_BUY_AMOUNT", createParameters(command))
}

func (client *UserClient) CancelSetBuyRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("GET", "CANCEL_SET_BUY", createParameters(command))
}

func (client *UserClient) SetBuyTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "SET_BUY_TRIGGER", createParameters(command))
}

func (client *UserClient) SetSellAmountRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "SET_SELL_AMOUNT", createParameters(command))
}

func (client *UserClient) CancelSetSellRequest(userid string, stockSymbol string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("GET", "CANCEL_SET_SELL", createParameters(command))
}

func (client *UserClient) SetSellTriggerRequest(userid string, stockSymbol string, amount string) (int, string, error) {
	var command = commandParams{
		UserID:      userid,
		Amount:      amount,
		StockSymbol: stockSymbol,
	}
	return client.makeRequest("POST", "SET_SELL_TRIGGER", createParameters(command))
}

func (client *UserClient) DumplogRequest(userid string, filename string) (int, string, error) {
	var command = commandParams{
		UserID:   userid,
		Filename: filename,
	}
	return client.makeRequest("POST", "DUMPLOG", createParameters(command))
}

func (client *UserClient) DisplaySummaryRequest(userid string) (int, string, error) {
	var command = commandParams{
		UserID: userid,
	}
	return client.makeRequest("POST", "DISPLAY_SUMMARY", createParameters(command))
}
