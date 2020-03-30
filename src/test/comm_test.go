package main

import (
	"net"
	"sync"
	"testing"
	"time"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
)

func TestSendRequestRetry(t *testing.T) {
	security.InitCryptoKey()
	payload := "test payload"
	address := ":6000"

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		listener, _ := net.Listen("tcp", address)
		_ = listener.Close()

		listener, _ = net.Listen("tcp", address)
		_ = listener.Close()

		listener, _ = net.Listen("tcp", address)
		serverConn, err := listener.Accept()
		if err != nil {
			t.Error("Comm package did not send 3 requests")
		}
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

func TestSendRequestNoRetry(t *testing.T) {
	security.InitCryptoKey()
	payload := "test payload"
	address := ":6000"

	timer1 := time.NewTimer(5 * time.Second)

	go func() {
		listener, _ := net.Listen("tcp", address)
		serverConn, err := listener.Accept()
		if err != nil {
			t.Error("Comm package did not send request")
		}
		serverConn.Close()
		<-timer1.C
		t.Error("Client should have immediately errored")
	}()

	_, _, err := lib.ClientSendRequest(address, payload)
	if err == nil {
		t.Error("Somehow did not error")
	}
}
