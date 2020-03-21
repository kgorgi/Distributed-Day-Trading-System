package main

import (
	"encoding/json"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/serverurls"
)

// TransactionClient client for transaction server
type TransactionClient struct{}

// SendCommand send command to transaction server
func (transactionClient *TransactionClient) SendCommand(command map[string]string) (int, string, error) {
	var err error
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return lib.StatusSystemError, "", err
	}

	status, message, err := lib.ClientSendRequest(serverurls.Env.TransactionServer, string(commandJSON))

	return status, message, err
}
