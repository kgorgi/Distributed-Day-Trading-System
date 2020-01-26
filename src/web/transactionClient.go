package main

import (
	"encoding/json"
	"extremeWorkload.com/daytrader/lib"
	"fmt"
	"net"
)

// TransactionClient client for transaction server
type TransactionClient struct {
	Network          string
	RemoteAddress    string
	ConnectionStatus string
	Socket           net.Conn
}

// ConnectSocket creates a socket connection to remote address
func (transactionClient *TransactionClient) ConnectSocket() (net.Conn, error) {
	var err error
	transactionClient.Socket, err = net.Dial(transactionClient.Network, transactionClient.RemoteAddress)
	if err != nil {
		fmt.Println(err)
	}
	return transactionClient.Socket, err
}

func (transactionClient TransactionClient) sendCommand(command map[string]string) (int, string, error) {
	commandJSON, err := json.Marshal(command)

	if err != nil {
		fmt.Println(err)
	}

	status, message, err2 := lib.ClientSendRequest(transactionClient.Socket, string(commandJSON))

	//handle reconnect here
	if err2 != nil {
		fmt.Println(err2)
	}
	return status, message, err
}
