package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

const maxRandomCents = 5000

func main() {
	fmt.Println("Starting mocked legacy quote server...")

	ln, _ := net.Listen("tcp", ":4443")

	fmt.Println("Started mocked legacy quote server on port: 4443")

	rand.Seed(time.Now().UnixNano())
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

	var amount = 5.00

	if !lib.DebuggingEnabled {
		amount = float64(rand.Int31n(maxRandomCents+1)) / float64(100)
	}

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
