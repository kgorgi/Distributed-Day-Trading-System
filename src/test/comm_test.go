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

	server, client := net.Pipe()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		message, _ := lib.ServerReceiveRequest(server)
		if message != payload {
			t.Errorf("message did not match payload:%s!=%s\n", message, payload)
		}
		lib.ServerSendResponse(server, lib.StatusSystemError, payload)

		message, _ = lib.ServerReceiveRequest(server)
		if message != payload {
			t.Errorf("message did not match payload:%s!=%s\n", message, payload)
		}
		lib.ServerSendResponse(server, lib.StatusSystemError, payload)

		message, _ = lib.ServerReceiveRequest(server)
		if message != payload {
			t.Errorf("message did not match payload:%s!=%s\n", message, payload)
		}
		lib.ServerSendResponse(server, lib.StatusOk, payload)
		wg.Done()
	}()

	status, message, _ := lib.ClientSendRequest(client, payload)
	if status != lib.StatusOk || message != payload {
		t.Error("Something went awry in the response")
	}
	wg.Wait()
}
