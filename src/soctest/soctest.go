package main

import (
	"fmt"
	"net"

	"extremeWorkload.com/daytrader/lib"
	commands "extremeWorkload.com/daytrader/lib/audit"
)

func main() {
	fmt.Println("Establishing Connection")

	conn, err := net.Dial("tcp", "localhost:5000")
	if err != nil {
		return
	}

	fmt.Println("Connection accepted")

	payload := commands.CreateLogSystemEvent("stest", 1, "test")

	status, _, _ := lib.ClientSendRequest(conn, payload)

	fmt.Println(status)
}
