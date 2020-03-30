package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
)

func main() {
	fmt.Println("Starting mocked legacy quote server...")

	ln, _ := net.Listen("tcp", ":4443")

	fmt.Println("Started mocked legacy quote server on port: 4443")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	lib.Debugln("Connection Established")
	rawRequest, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		lib.Errorln("Failed to get request " + err.Error())
		conn.Close()
		return
	}

	rawRequest = strings.TrimRight(rawRequest, "\n")
	data := strings.Split(rawRequest, ",")
	if len(data) != 2 {
		lib.Errorln("Invalid request " + rawRequest)
		conn.Close()
		return
	}

	amount := 5.00
	stockSymbol := data[0]
	userID := data[1]
	timestamp := lib.GetUnixTimestamp()
	cryptoKey := "4DxwFafID/pjlWjAUpX+1xpHLvP6EzX7BWeZVUjq2Ev9RT0CDnd8mQ=="
	result := fmt.Sprintf("%f,%s,%s,%d,%s\n", amount, stockSymbol, userID, timestamp, cryptoKey)

	_, err = conn.Write([]byte(result))
	if err != nil {
		lib.Errorln("Failed to write response to:  " + rawRequest + " " + err.Error())
	}

	conn.Close()

	lib.Debugln("Connection Closed")
}
