package dataclient

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"extremeWorkload.com/daytrader/lib/resolveurl"
)

var (
	// ErrNotFound is returned when a user or trigger is not found in the database.
	ErrNotFound = errors.New("not found")
)

// CreateUser takes a user struct and creates a user in the database
func CreateUser(user modelsdata.User) error {
	userBytes, jsonErr := json.Marshal(user)
	if jsonErr != nil {
		return jsonErr
	}
	userJSON := string(userBytes)

	payload := "CREATE_USER|" + userJSON
	_, _, err := sendRequest(payload)
	return err
}

// ReadUsers reads all users from the database
func ReadUsers() ([]modelsdata.User, error) {
	users := make([]modelsdata.User, 0)

	payload := "READ_USERS"
	_, message, err := sendRequest(payload)
	if err != nil {
		return users, err
	}

	jsonErr := json.Unmarshal([]byte(message), &users)
	if jsonErr != nil {
		return users, jsonErr
	}

	return users, nil
}

// ReadUser takes userID and reads a user from the database
func ReadUser(userID string) (modelsdata.User, error) {
	payload := "READ_USER|" + userID
	_, message, err := sendRequest(payload)
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

// UpdateUser takes in a user struct and updates the corresponding user in the database
func UpdateUser(user modelsdata.User) error {
	userBytes, jsonErr := json.Marshal(user)
	if jsonErr != nil {
		return jsonErr
	}
	userJSON := string(userBytes)

	payload := "UPDATE_USER|" + userJSON
	_, _, err := sendRequest(payload)
	return err
}

// DeleteUser takes a userID and deletes the corresponding user from the database
func DeleteUser(userID string) error {
	payload := "DELETE_USER|" + userID
	_, _, err := sendRequest(payload)
	return err
}

// CreateTrigger takes a trigger struct and creates a trigger in the database
func CreateTrigger(trigger modelsdata.Trigger) error {
	triggerBytes, jsonErr := json.Marshal(trigger)
	if jsonErr != nil {
		return jsonErr
	}
	triggerJSON := string(triggerBytes)

	payload := "CREATE_TRIGGER|" + triggerJSON
	_, _, err := sendRequest(payload)
	return err
}

// ReadTriggers reads all triggers from the database
func ReadTriggers() ([]modelsdata.Trigger, error) {
	triggers := make([]modelsdata.Trigger, 0)

	payload := "READ_TRIGGERS"
	_, message, err := sendRequest(payload)
	if err != nil {
		return triggers, err
	}

	jsonErr := json.Unmarshal([]byte(message), &triggers)
	if jsonErr != nil {
		return triggers, jsonErr
	}

	return triggers, nil
}

// ReadTriggersByUser takes a userID and reads all assosiated triggers from the database
func ReadTriggersByUser(userID string) ([]modelsdata.Trigger, error) {
	triggers := make([]modelsdata.Trigger, 0)

	payload := "READ_TRIGGERS|" + userID
	_, message, err := sendRequest(payload)
	if err != nil {
		return triggers, err
	}

	jsonErr := json.Unmarshal([]byte(message), &triggers)
	if jsonErr != nil {
		return triggers, jsonErr
	}

	return triggers, nil
}

// ReadTrigger takes the primary key attributes for a trigger and reads a trigger from the database
func ReadTrigger(userID string, stockName string, isSell bool) (modelsdata.Trigger, error) {
	payload := "READ_TRIGGER|" + userID + "|" + stockName + "|" + strconv.FormatBool(isSell)
	_, message, err := sendRequest(payload)
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

// UpdateTrigger takes a trigger struct and updates the corresponding trigger in the database
func UpdateTrigger(trigger modelsdata.Trigger) error {
	triggerBytes, jsonErr := json.Marshal(trigger)
	if jsonErr != nil {
		return jsonErr
	}
	triggerJSON := string(triggerBytes)

	payload := "UPDATE_TRIGGER|" + triggerJSON
	_, _, err := sendRequest(payload)
	return err
}

// DeleteTrigger takes the primary key attributes of a trigger and deletes the corresponding trigger in the database
func DeleteTrigger(userID string, stockName string, isSell bool) error {

	payload := "DELETE_TRIGGER|" + userID + "|" + stockName + "|" + strconv.FormatBool(isSell)
	_, _, err := sendRequest(payload)
	return err
}

func sendRequest(payload string) (int, string, error) {
	//connect to data server
	conn, err := net.Dial("tcp", resolveurl.DataServerAddress())
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
		if status == lib.StatusNotFound {
			return status, message, ErrNotFound
		}
		return status, message, errors.New("Not ok, status: " + strconv.Itoa(status))
	}

	return status, message, nil
}
