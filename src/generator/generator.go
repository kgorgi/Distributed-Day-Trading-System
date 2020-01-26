package main

import (
	"fmt"
	"net"
	"extremeWorkload.com/daytrader/lib"
)
func main() {
	fmt.Println("Establishing Connection")
	conn, err := net.Dial("tcp", "localhost:5001")
	if err != nil {
		return
	}
	fmt.Println("Connection accepted")
	// userJson := `{"command_id": "serverTest", "cents": 66, "investments": []}`
	// payloadc := "CREATE_USER|" + userJson;
	// cstatus, cmessage, _ := lib.ClientSendRequest(conn, payloadc);
	// fmt.Println(cmessage)
	// fmt.Println(cstatus)

	userJson := `{"command_id": "serverTest", "cents": 1738, "investments": [{"stock": "ABC", "amount": 68}]}`
	payloadc := "UPDATE_USER|" + userJson;
	cstatus, cmessage, _ := lib.ClientSendRequest(conn, payloadc);
	fmt.Println(cmessage)
	fmt.Println(cstatus)


	payload := "READ_USERS"
	status, message, _ := lib.ClientSendRequest(conn, payload)
	fmt.Println(message)
	fmt.Println(status)
}
