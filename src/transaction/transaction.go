package main

import (
	"fmt"
	"net"
	"extremeWorkload.com/daytrader/lib"
)

var dataConn net.Conn

func handleConnection(conn net.Conn) {
	for {
		// will listen for message to process ending in newline (\n)
		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			conn.Close()
			break
		}

		userJson := `{"command_id": "serverTest", "cents": 66, "investments": []}`
		payloadc := "CREATE_USER|" + userJson;
		cstatus, cmessage, _ := lib.ClientSendRequest(dataConn, payloadc);
		fmt.Println(cmessage)
		fmt.Println(cstatus)
		
		// userJson := `{"command_id": "serverTest", "cents": 1738, "investments": [{"stock": "ABC", "amount": 68}]}`
		// payloadc := "UPDATE_USER|" + userJson;
		// cstatus, cmessage, _ := lib.ClientSendRequest(conn, payloadc);
		// fmt.Println(cmessage)
		// fmt.Println(cstatus)
		
		payload2 := "READ_USERS"
		status, message, _ := lib.ClientSendRequest(dataConn, payload2)
		fmt.Println(message)
		fmt.Println(status)

		//data := strings.Split(payload, "|")
		fmt.Println("received: " + payload)
		lib.ServerSendResponse(conn, lib.StatusOk, payload)
	}

	fmt.Println("closed client")
}

func main() {
	fmt.Println("Establishing Connection")
	var err error
    dataConn, err = net.Dial("tcp", "data-server:5001")
    if err != nil {
        return
    }
	fmt.Println("Connection accepted")

	fmt.Println("launching server...")

	ln, _ := net.Listen("tcp", ":5000")

	for {
		conn, _ := ln.Accept()
		fmt.Println("new client accepted")
		go handleConnection(conn)
	}
}
