package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

var cryptokey = "KO2gt9eJ+aJrvTjHuBGdzTKUaKchO2piV6WILGdt4t5DnL+oBkIqJA=="

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
	request, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return
	}

	data := strings.Split(request, ",")
	stockSymbol := data[0]
	userid := strings.Trim(data[1], "\n")
	timestamp := uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
	response := "5.00," + stockSymbol + "," + userid + "," + strconv.FormatUint(timestamp, 10) + "," + cryptokey + "\n"
	conn.Write([]byte(response))
	conn.Close()
	lib.Debugln("Connection Closed")
}
