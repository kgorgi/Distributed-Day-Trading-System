package dataclient

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	auditclient "extremeWorkload.com/daytrader/lib/audit"

	"extremeWorkload.com/daytrader/lib/serverurls"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
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

// UpdateUser increments/decrements a users stocks and money
func UpdateUser(userID string, stock string, amount int, cents int, auditClient *auditclient.AuditClient) error {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.UpdateUserCommand{
			UserID:      userID,
			Stock:       stock,
			StockAmount: amount,
			Cents:       cents,
		},
	)

	if jsonErr != nil {
		return jsonErr
	}

	payload := "UPDATE_USER|" + string(commandBytes)
	_, _, err := sendRequest(payload)

	if err == nil && cents != 0 {
		auditClient.LogAccountTransaction(userID, int64(cents))
	}

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
	commandBytes, jsonErr := json.Marshal(
		modelsdata.ChooseTriggerCommand{
			UserID: userID,
			Stock:  stockName,
			IsSell: isSell,
		},
	)

	if jsonErr != nil {
		return modelsdata.Trigger{}, jsonErr
	}

	payload := "READ_TRIGGER|" + string(commandBytes)
	_, message, err := sendRequest(payload)
	if err != nil {
		return modelsdata.Trigger{}, err
	}

	var trigger modelsdata.Trigger
	jsonErr = json.Unmarshal([]byte(message), &trigger)
	if jsonErr != nil {
		return modelsdata.Trigger{}, jsonErr
	}

	return trigger, nil
}

// DeleteTrigger takes the primary key attributes of a trigger and deletes the corresponding trigger in the database
// it returns the successfully deleted trigger
func DeleteTrigger(userID string, stockName string, isSell bool) (modelsdata.Trigger, error) {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.ChooseTriggerCommand{
			UserID: userID,
			Stock:  stockName,
			IsSell: isSell,
		},
	)

	if jsonErr != nil {
		return modelsdata.Trigger{}, jsonErr
	}

	payload := "DELETE_TRIGGER|" + string(commandBytes)
	_, message, err := sendRequest(payload)

	if err != nil {
		return modelsdata.Trigger{}, err
	}

	var deletedTrigger modelsdata.Trigger
	jsonErr = json.Unmarshal([]byte(message), &deletedTrigger)
	if jsonErr != nil {
		return modelsdata.Trigger{}, jsonErr
	}

	return deletedTrigger, nil
}

// UpdateTriggerPrice updates the price at which a trigger will fire for its stock
func UpdateTriggerPrice(userID string, stock string, isSell bool, price uint64) error {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.UpdateTriggerPriceCommand{
			UserID: userID,
			Stock:  stock,
			IsSell: isSell,
			Price:  price,
		},
	)

	if jsonErr != nil {
		return jsonErr
	}

	payload := "UPDATE_TRIGGER_PRICE|" + string(commandBytes)
	_, _, err := sendRequest(payload)
	return err
}

// UpdateTriggerAmount updates the amount of cents worth of a stock a trigger will buy or sell if it's price condition is met
func UpdateTriggerAmount(userID string, stock string, isSell bool, amount uint64) error {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.UpdateTriggerAmountCommand{
			UserID: userID,
			Stock:  stock,
			IsSell: isSell,
			Amount: amount,
		},
	)

	if jsonErr != nil {
		return jsonErr
	}

	payload := "UPDATE_TRIGGER_AMOUNT|" + string(commandBytes)
	_, _, err := sendRequest(payload)
	return err
}

// PushUserBuy adds a buy to a users stack
func PushUserBuy(userID string, stock string, cents uint64, numStock uint64) error {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.PushUserReserveCommand{
			UserID:   userID,
			Stock:    stock,
			Cents:    cents,
			NumStock: numStock,
		},
	)

	if jsonErr != nil {
		return jsonErr
	}

	payload := "PUSH_USER_BUY|" + string(commandBytes)
	_, _, err := sendRequest(payload)
	return err
}

// PopUserBuy pops a buy from the users stack, it will return the not found error
// if either the user is not found, or they have no valid buys in their stack
func PopUserBuy(userID string) (modelsdata.Reserve, error) {
	payload := "POP_USER_BUY|" + userID
	_, message, err := sendRequest(payload)
	if err != nil {
		return modelsdata.Reserve{}, err
	}

	var buyReserve modelsdata.Reserve
	jsonErr := json.Unmarshal([]byte(message), &buyReserve)
	if jsonErr != nil {
		return modelsdata.Reserve{}, jsonErr
	}

	return buyReserve, nil
}

// PushUserSell adds a sell to a users stack
func PushUserSell(userID string, stock string, cents uint64, numStock uint64) error {
	commandBytes, jsonErr := json.Marshal(
		modelsdata.PushUserReserveCommand{
			UserID:   userID,
			Stock:    stock,
			Cents:    cents,
			NumStock: numStock,
		},
	)

	if jsonErr != nil {
		return jsonErr
	}

	payload := "PUSH_USER_SELL|" + string(commandBytes)
	_, _, err := sendRequest(payload)
	return err
}

// PopUserSell pops a sell from the users stack, it will return the not found error
// if either the user is not found, or they have no valid sells in their stack
func PopUserSell(userID string) (modelsdata.Reserve, error) {
	payload := "POP_USER_SELL|" + userID
	_, message, err := sendRequest(payload)
	if err != nil {
		return modelsdata.Reserve{}, err
	}

	var sellReserve modelsdata.Reserve
	jsonErr := json.Unmarshal([]byte(message), &sellReserve)
	if jsonErr != nil {
		return modelsdata.Reserve{}, jsonErr
	}

	return sellReserve, nil
}

func sendRequest(payload string) (int, string, error) {
	// Send Payload
	status, message, err := lib.ClientSendRequest(serverurls.Env.DataServer, payload)

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
