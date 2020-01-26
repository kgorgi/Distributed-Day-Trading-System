package main

import (
	"fmt"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
)

var dataConn net.Conn

func isUseridValid(userid string) (bool, error) {
	if true {
		return true, nil
	}
	return false, &InvalidData{userid + " is not a valid userid"}
}

func isAmountValid(amount string) (bool, error) {
	return true, nil
}

func isUserExist(userid string) (bool, error) {
	return true, nil
}

func createUser(userid string) (bool, error) {
	return true, nil
}

func addAmount(userid string, amount string) (bool, error) {
	return true, nil
}

func processTransaction(transactionRequest string) (int, string) {
	// COMMAND, USERNAME, P1, ...
	transactionCommand := strings.Split(transactionRequest, ",")
	userid := transactionCommand[1]
	isValid, err := isUseridValid(userid)
	if !isValid {
		return lib.StatusUserError, err.Error()
	}

	switch strings.ToUpper(transactionCommand[0]) {
	case "ADD":
		// check if amount is valid
		isValid, err = isAmountValid(transactionCommand[2])
		if !isValid {
			return lib.StatusUserError, err.Error()
		}
		// if not userid create user
		isValid, err = isUserExist(transactionCommand[2])
		if !isValid {
			createUser(userid)
		}
		// add amount
		addAmount(userid, transactionCommand[2])
		return lib.StatusOk, "add processed"
	}

	return lib.StatusUserError, "command doesn't exist"
}

func handleWebConnection(conn net.Conn) {
	for {
		payload, err := lib.ServerReceiveRequest(conn)
		if err != nil {
			conn.Close()
			break
		}

		// e2e test
		// userJson := `{"command_id": "serverTest", "cents": 66, "investments": []}`
		// payloadc := "CREATE_USER|" + userJson
		// cstatus, cmessage, _ := lib.ClientSendRequest(dataConn, payloadc)
		// fmt.Println(cmessage)
		// fmt.Println(cstatus)

		// // userJson := `{"command_id": "serverTest", "cents": 1738, "investments": [{"stock": "ABC", "amount": 68}]}`
		// // payloadc := "UPDATE_USER|" + userJson;
		// // cstatus, cmessage, _ := lib.ClientSendRequest(conn, payloadc);
		// // fmt.Println(cmessage)
		// // fmt.Println(cstatus)

		// payload2 := "READ_USERS"
		// status, message, _ := lib.ClientSendRequest(dataConn, payload2)
		// fmt.Println(message)
		// fmt.Println(status)

		//data := strings.Split(payload, "|")

		fmt.Println("received: " + payload)
		status, message := processTransaction(payload)
		lib.ServerSendResponse(conn, status, message)
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
		go handleWebConnection(conn)
	}
}
