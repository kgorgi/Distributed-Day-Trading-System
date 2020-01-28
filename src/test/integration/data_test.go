package main

import (
	"fmt"
	"testing"
	dataclient "extremeWorkload.com/daytrader/lib/data"
)

func TestDataServer(t *testing.T) {
	var dataClient = dataclient.DataClient{}

	createErr := dataClient.CreateUser(`{"command_id": "serverTest", "cents": 66, "investments": []}`);
	if createErr != nil {
		fmt.Println(createErr);
	}
	
	users, readErr := dataClient.ReadUsers();
	if readErr != nil {
		fmt.Println(readErr)
	}

	fmt.Println(users);
}