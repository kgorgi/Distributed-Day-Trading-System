package dataclient

import (
	"log"
	"net"
	"errors"
	"strconv"
	"encoding/json"
	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/models/data"
	"extremeWorkload.com/daytrader/lib/url"
)


type DataClient struct {}

func (client *DataClient) CreateUser(user modelsdata.User) error{
	userBytes, jsonErr := json.Marshal(user);
	if jsonErr != nil {
		return jsonErr;
	}
	userJSON := string(userBytes)

	payload := "CREATE_USER|" + userJSON
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) ReadUsers() ([]modelsdata.User, error) {
	users := make([]modelsdata.User, 0)

	payload := "READ_USERS"
	_, message, err := client.sendRequest(payload);
	if err != nil {
		return users, err
	}

	jsonErr := json.Unmarshal([]byte(message), &users);
	if jsonErr != nil {
		return users, jsonErr;
	}

	return users, nil
}

func (client *DataClient) ReadUser(userID string) (modelsdata.User, error) {
	payload := "READ_USER|" + userID;
	_, message, err := client.sendRequest(payload);
	if err != nil {
		return modelsdata.User{}, err
	}

	var user modelsdata.User
	jsonErr := json.Unmarshal([]byte(message), &user)
	if jsonErr != nil {
		return modelsdata.User{}, err
	}

	return user, nil
}

func (client *DataClient) UpdateUser(user modelsdata.User) error {
	userBytes, jsonErr := json.Marshal(user);
	if jsonErr != nil {
		return jsonErr;
	}
	userJSON := string(userBytes)

	payload := "UPDATE_USER|" + userJSON;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) DeleteUser(userID string) error {
	payload := "DELETE_USER|" + userID;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) CreateTrigger(trigger modelsdata.Trigger) error {
	triggerBytes, jsonErr := json.Marshal(trigger);
	if jsonErr != nil {
		return jsonErr;
	}
	triggerJSON := string(triggerBytes)

	payload := "CREATE_TRIGGER|" + triggerJSON;
	_, _, err := client.sendRequest(payload);
	return err;
}

func (client *DataClient) ReadTriggers() ([]modelsdata.Trigger, error) {
	triggers := make([]modelsdata.Trigger, 0)

	payload := "READ_TRIGGERS"
	_, message, err := client.sendRequest(payload)
	if err != nil {
		return triggers, err
	}

	jsonErr := json.Unmarshal([]byte(message), &triggers)
	if jsonErr != nil {
		return triggers, jsonErr
	}

	return triggers, nil
}

func (client *DataClient) ReadTrigger(userID string, stockName string) (modelsdata.Trigger, error) {
	payload := "READ_TRIGGER|" + userID + "|" + stockName
	_, message, err := client.sendRequest(payload);
	if err != nil {
		return modelsdata.Trigger{}, err
	}

	var trigger modelsdata.Trigger
	jsonErr := json.Unmarshal([]byte(message), &trigger)
	if jsonErr != nil {
		return modelsdata.Trigger{}, jsonErr
	}

	return trigger, nil
}

func (client *DataClient) UpdateTrigger(trigger modelsdata.Trigger) error {
	triggerBytes, jsonErr := json.Marshal(trigger);
	if jsonErr != nil {
		return jsonErr;
	}
	triggerJSON := string(triggerBytes)

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
	conn, err := net.Dial("tcp", url.ResolveDataServerAddress())
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
		return status, message, errors.New("Not Ok, status: " + string(status));
	}

	return status, message, nil
}