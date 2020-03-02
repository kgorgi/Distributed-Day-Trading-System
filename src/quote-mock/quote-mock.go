package main

import (
	"bufio"
	"fmt"
	"net"

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
	bufio.NewReader(conn).ReadString('\n')
	conn.Write([]byte("5.00,ABC,quoteMock,123456,KEY\n"))
	conn.Close()
	lib.Debugln("Connection Closed")
}
