package main

import (
	"bufio"
	"net"
)

func main() {

	ln, _ := net.Listen("tcp", ":4443")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	bufio.NewReader(conn).ReadString('\n')
	conn.Write([]byte("5.00,ABC,quoteMock,123456,KEY\n"))
	conn.Close()
}
