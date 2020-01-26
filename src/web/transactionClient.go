package main

import (
	"fmt"
	"net"

	"extremeWorkload.com/daytrader/lib"
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
	return transactionClient.Socket, err
}

func (transactionClient TransactionClient) sendCommand(command string) (int, string, error) {
	status, message, err := lib.ClientSendRequest(transactionClient.Socket, command)

	//handle reconnect here
	if err != nil {
		fmt.Print(err)
	}
	return status, message, err
}
