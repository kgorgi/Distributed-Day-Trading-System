package main

import (
	"encoding/json"
	"net"

	"extremeWorkload.com/daytrader/lib"
)

const transactionServerDockerAddress = "transaction-server:5000"

// const transactionServerDockerAddress = ":5000"

// TransactionClient client for transaction server
type TransactionClient struct{}

// SendCommand send command to transaction server
func (transactionClient *TransactionClient) SendCommand(command map[string]string) (int, string, error) {
	var err error
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return lib.StatusSystemError, "", err
	}

	conn, err := net.Dial("tcp", ":5000")
	if err != nil {
		return lib.StatusSystemError, "", err
	}

	status, message, err := lib.ClientSendRequest(conn, string(commandJSON))

	conn.Close()

	return status, message, err
}
