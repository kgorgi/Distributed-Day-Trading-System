package main

import (
	"net"
	"sync"
	"testing"

	"extremeWorkload.com/daytrader/lib"
)

func TestHealthLocal(t *testing.T) {
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Error("Failed to setup test listener")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		server, _ := listener.Accept()
		isHealthCheck, err := lib.ServerReceiveHealthCheck(server)
		if err != nil {
			t.Error("Server read failed: " + err.Error())
		}
		if !isHealthCheck {
			t.Errorf("Unexpected output: %t", isHealthCheck)
		}
		server.Close()
		wg.Done()
	}()

	err = TCPHealthCheck(":9999")
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	wg.Wait()

}
