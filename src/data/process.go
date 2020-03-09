package main

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/mongo"
)

func generateIsSellBool(isSellString string) bool {
	return isSellString == "true"
}

func processCommand(conn net.Conn, client *mongo.Client, payload string) {
	data := strings.Split(payload, "|")
	command := data[0]
	switch command {
	case "CREATE_USER":
		userJson := data[1]
		var newUser modelsdata.User
		jsonErr := json.Unmarshal([]byte(userJson), &newUser)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		createErr := createUser(client, newUser)
		if createErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)
	case "READ_USER":
		commandID := data[1]
		user, readErr := readUser(client, commandID)
		if readErr != nil {
			status := lib.StatusSystemError
			if readErr == mongo.ErrNoDocuments {
				status = lib.StatusNotFound
			}
			lib.ServerSendResponse(conn, status, readErr.Error())
			break
		}

		userBytes, jsonErr := json.Marshal(user)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(userBytes))
	case "READ_USERS":
		users, readErr := readUsers(client)
		if readErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
			break
		}

		usersBytes, jsonErr := json.Marshal(users)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(usersBytes))
	case "UPDATE_USER":
		userCommandID := data[1]
		stock := data[2]
		amount := data[3]
		cents := data[4]

		//If no stock should be added or removed
		if stock == "" || amount == "" || amount == "0" {
			centsInt, conversionErr := strconv.Atoi(cents)
			if conversionErr != nil {
				lib.ServerSendResponse(conn, lib.StatusUserError, "Cents and Amount must be integers")
				break
			}

			updateErr := updateCents(client, userCommandID, centsInt)
			if updateErr != nil {
				lib.ServerSendResponse(conn, lib.StatusUserError, "The specified user does not exist, or they do not have the specified amount of money")
			}

			lib.ServerSendOKResponse(conn)
		}

		//Otherwise we need to modify the users stock and money
		centsInt, conversionErr1 := strconv.Atoi(cents) //TODO: command structs
		amountInt, conversionErr2 := strconv.Atoi(amount)
		if conversionErr1 != nil || conversionErr2 != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Cents and Amount must be integers")
			return
		}

		updateErr := updateStockAndCents(client, userCommandID, stock, amountInt, centsInt)
		if updateErr == ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Either the user does not exist, or they do not have sufficient stock or funds to remove")
			return
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			return
		}

		lib.ServerSendOKResponse(conn)

	case "CREATE_TRIGGER":
		triggerJson := data[1]
		var newTrigger modelsdata.Trigger
		jsonErr := json.Unmarshal([]byte(triggerJson), &newTrigger)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		createErr := createTrigger(client, newTrigger)
		if createErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)
	case "READ_TRIGGER":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]

		trigger, readErr := readTrigger(client, userCommandID, stock, generateIsSellBool(isSellString))
		if readErr != nil {
			status := lib.StatusSystemError
			if readErr == mongo.ErrNoDocuments {
				status = lib.StatusNotFound
			}
			lib.ServerSendResponse(conn, status, readErr.Error())
			break
		}

		triggerBytes, jsonErr := json.Marshal(trigger)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(triggerBytes))
	case "READ_TRIGGERS":
		var triggers []modelsdata.Trigger
		var readErr error

		// if the length of data is greater than 1 that means user_command_ID is included in the command
		if len(data) > 1 {
			triggers, readErr = readTriggersByUser(client, data[1])
		} else {
			triggers, readErr = readTriggers(client)
		}

		if readErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
			break
		}

		triggersBytes, jsonErr := json.Marshal(triggers)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(triggersBytes))
	case "UPDATE_TRIGGER":
		triggerJson := data[1]
		var triggerUpdate modelsdata.Trigger
		jsonErr := json.Unmarshal([]byte(triggerJson), &triggerUpdate)

		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		updateErr := updateTrigger(client, triggerUpdate)
		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)
	case "DELETE_TRIGGER":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]

		deletedTrigger, deleteErr := deleteTrigger(client, userCommandID, stock, generateIsSellBool(isSellString))
		if deleteErr == ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, deleteErr.Error())
			return
		}

		if deleteErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
			break
		}

		triggerBytes, jsonErr := json.Marshal(deletedTrigger)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(triggerBytes))

	case "UPDATE_TRIGGER_PRICE":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]
		price := data[4]

		priceInt, conversionErr := strconv.ParseUint(price, 10, 64)
		if conversionErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Price must be an unsigned integer")
		}

		updateErr := updateTriggerPrice(client, userCommandID, stock, generateIsSellBool(isSellString), priceInt)
		if updateErr == ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified Trigger does not exist, or its amount is less than the specified price")
			break
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)

	case "UPDATE_TRIGGER_AMOUNT":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]
		amount := data[4]

		amountInt, conversionErr := strconv.ParseUint(amount, 10, 64)
		if conversionErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Amount must be an unsigned integer")
		}

		updateErr := updateTriggerAmount(client, userCommandID, stock, generateIsSellBool(isSellString), amountInt)
		if updateErr == ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified Trigger does not exist")
			break
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)

	case "SET_TRIGGER_AMOUNT":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]
		amount := data[4]
		transactionNumber := data[5]

		amountInt, conversionErr1 := strconv.ParseUint(amount, 10, 64)
		transactionNumberInt, conversionErr2 := strconv.ParseUint(transactionNumber, 10, 64)
		if conversionErr1 != nil && conversionErr2 != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Amount must be an unsigned integer")
		}

		oldTrigger, setErr := setTriggerAmount(client, userCommandID, stock, generateIsSellBool(isSellString), amountInt, transactionNumberInt)
		if setErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, setErr.Error())
			break
		}

		// If the trigger and the error are nil, that means a new trigger was created and there is no old
		// Trigger to return to the client
		if oldTrigger == nil {
			lib.ServerSendOKResponse(conn)
			break
		}

		triggerBytes, jsonErr := json.Marshal(*oldTrigger)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(triggerBytes))

	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Data Server Command")
	}
}
