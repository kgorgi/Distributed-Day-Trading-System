package main

import (
	"net"
	"sync"
	"testing"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
)

func TestSendRequestRetry(t *testing.T) {
	security.InitCryptoKey()
	payload := "test payload"
	address := ":6000"

	var wg sync.WaitGroup

	listener, err := net.Listen("tcp", address)
	defer listener.Close()
	if err != nil {
		t.Error(err.Error())
	}

	wg.Add(1)
	go func() {
		serverConn, _ := listener.Accept()
		serverConn.Close()

		serverConn, _ = listener.Accept()
		serverConn.Close()

		serverConn, _ = listener.Accept()
		lib.ServerReceiveRequest(serverConn)
		lib.ServerSendOKResponse(serverConn)
		serverConn.Close()

		wg.Done()
	}()
	status, _, _ := lib.ClientSendRequest(address, payload)
	if status != lib.StatusOk {
		t.Error("Something went awry in the response")
	}
	wg.Wait()
}
