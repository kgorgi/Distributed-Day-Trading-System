package main

import (
	"net"
	"os"
	"sync"
	"testing"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
)

func TestHealthLocal(t *testing.T) {
	security.InitCryptoKey()
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Error("Failed to setup test listener")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		os.Setenv("CHECK_TRIGGERS", "yes")
		server, _ := listener.Accept()
		payload, err := lib.ServerReceiveRequest(server)
		if err != nil {
			t.Error("Server read failed: " + err.Error())
		}
		if payload != lib.HealthCheck {
			t.Errorf("Unexpected output: %s", payload)
		}
		lib.ServerSendHealthResponse(server, lib.HealthStatusUp)
		server.Close()
		wg.Done()
	}()

	message, err := TCPHealthCheck(":9999")
	if err != nil {
		t.Error("Health Check failed: " + err.Error())
	}
	if message != lib.HealthStatusUp {
		t.Error("Recieved unexpected response: " + message)
	}
	wg.Wait()

}
