package main

import (
	"fmt"
	"net"

	"extremeWorkload.com/daytrader/lib"
)

func handleConnection(conn net.Conn) {
	for {
		// will listen for message to process ending in newline (\n)
		message, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			conn.Close()
			break
		}
		fmt.Println("received: " + message)
		lib.ServerSendResponse(conn, lib.StatusOk, message)
	}

	fmt.Println("closed client")
}

func main() {

	fmt.Println("launching server...")

	ln, _ := net.Listen("tcp", ":8081")

	for {
		conn, _ := ln.Accept()
		fmt.Println("new client accepted")
		go handleConnection(conn)
	}
}
