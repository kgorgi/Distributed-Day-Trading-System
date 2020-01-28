package dataclient

import (
	"log"
	"net"
	"strconv"
	"extremeWorkload.com/daytrader/lib"
)

const dataServerDockerAddress = "data-server:5001"
const dataServerLocalAddress = "localhost:5001"

type DataClient struct {}

func (client *DataClient) CreateUser(userJSON string) error{
	payload := "CREATE_USER|" + userJSON
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) ReadUsers() (string, error) {
	payload := "READ_USERS"
	_, message, err := client.sendRequest(payload);
	return message, err
}

func (client *DataClient) ReadUser(userID string) (string, error) {
	payload := "READ_USER|" + userID;
	_, message, err := client.sendRequest(payload);
	return message, err;
}

func (client *DataClient) UpdateUser(userJSON string) error {
	payload := "UPDATE_USER|" + userJSON;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) DeleteUser(userID string) error {
	payload := "DELETE_USER|" + userID;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) CreateTrigger(triggerJSON string) error {
	payload := "CREATE_TRIGGER|" + triggerJSON;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) ReadTriggers() (string, error) {
	payload := "READ_TRIGGERS"
	_, message, err := client.sendRequest(payload)
	return message, err
}

func (client *DataClient) ReadTrigger(userID string, stockName string) (string, error) {
	payload := "READ_TRIGGER|" + userID + "|" + stockName
	_, message, err := client.sendRequest(payload);
	return message, err
}

func (client *DataClient) UpdateTrigger(triggerJSON string) error {
	payload := "UPDATE_TRIGGER|" + triggerJSON
	_, _, err := client.sendRequest(payload);
	return err
}

func (client *DataClient) DeleteTrigger(userID string, stockName string) error {
	payload := "DELETE_TRIGGER|" + userID + "|" + stockName
	_, _, err := client.sendRequest(payload)
	return err;
}

func (client *DataClient) sendRequest(payload string) (int, string, error) {
	//connect to data server
	conn, err := net.Dial("tcp", dataServerDockerAddress)
	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return -1, "", err
	}

	// Send Payload
	status, message, err := lib.ClientSendRequest(conn, payload)

	conn.Close()

	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return -1, "", err
	}

	if status != lib.StatusOk {
		log.Println("Response Error: Status " + strconv.Itoa(status) + " " + message)
		return status, message, nil
	}

	return status, message, nil
}