package main

import (
	"encoding/json"
	"extremeWorkload.com/daytrader/lib"
	"fmt"
	"net"
)

const transactionServerDockerAddress = "transaction-server:5000"

// const transactionServerDockerAddress = ":5000"

// TransactionClient client for transaction server
type TransactionClient struct{}

// SendCommand send command to transaction server
func (transactionClient TransactionClient) SendCommand(command map[string]string) (int, string, error) {
	var err error
	commandJSON, err := json.Marshal(command)

	if err != nil {
		fmt.Println(err)
	}
	conn, err := net.Dial("tcp", transactionServerDockerAddress)
	if err != nil {
		fmt.Println(err)
	}

	status, message, err2 := lib.ClientSendRequest(conn, string(commandJSON))

	conn.Close()

	//handle reconnect here
	if err2 != nil {
		fmt.Println(err2)
	}
	return status, message, err
}
