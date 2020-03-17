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

// TODO: Instead of having a massive switch statement write functions for handling each request
// Create some command structs, how I verify command inputs is pretty bad
// Better more detailed error handling
// Figure out a nice way to combine UPDATE_TRIGGER_AMOUNT and UPDATE_TRIGGER_PRICE into one command

func generateIsSellBool(isSellString string) bool {
	return isSellString == "true"
}

func processCommand(conn net.Conn, client *mongo.Client, payload string) {
	data := strings.Split(payload, "|")
	command := data[0]
	switch command {
	case "CREATE_USER":
		userJSON := data[1]
		var newUser modelsdata.User
		jsonErr := json.Unmarshal([]byte(userJSON), &newUser)
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
		if updateErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusUserError, "Either the user does not exist, or they do not have sufficient stock or funds to remove")
			return
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			return
		}

		lib.ServerSendOKResponse(conn)

	case "CREATE_TRIGGER":
		triggerJSON := data[1]
		var newTrigger modelsdata.Trigger
		jsonErr := json.Unmarshal([]byte(triggerJSON), &newTrigger)
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

	case "DELETE_TRIGGER":
		userCommandID := data[1]
		stock := data[2]
		isSellString := data[3]

		deletedTrigger, deleteErr := deleteTrigger(client, userCommandID, stock, generateIsSellBool(isSellString))
		if deleteErr == errNotFound {
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
		if updateErr == errNotFound {
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
		if updateErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified Trigger does not exist")
			break
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)

	case "PUSH_USER_BUY":
		userCommandID := data[1]
		stock := data[2]
		cents := data[3]
		numStock := data[4]

		numStockInt, conversionErr1 := strconv.ParseUint(numStock, 10, 64)
		centsInt, conversionErr2 := strconv.ParseUint(cents, 10, 64)
		if conversionErr1 != nil || conversionErr2 != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "cents and number of stocks must be unsigned integers")
			break
		}

		pushErr := pushUserReserve(client, userCommandID, stock, centsInt, numStockInt, false)
		if pushErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified user does not exist")
			break
		}

		if pushErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, pushErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)

	case "POP_USER_BUY":
		userCommandID := data[1]

		buy, popErr := popUserReserve(client, userCommandID, false)
		if popErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified user does not exist")
			break
		}

		if popErr == errEmptyStack {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "the buy stack is empty")
			break
		}

		if popErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, popErr.Error())
			break
		}

		buyBytes, jsonErr := json.Marshal(buy)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(buyBytes))

	case "PUSH_USER_SELL":
		userCommandID := data[1]
		stock := data[2]
		cents := data[3]
		numStock := data[4]

		numStockInt, conversionErr1 := strconv.ParseUint(numStock, 10, 64)
		centsInt, conversionErr2 := strconv.ParseUint(cents, 10, 64)
		if conversionErr1 != nil || conversionErr2 != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, "cents and number of stocks must be unsigned integers")
			break
		}

		pushErr := pushUserReserve(client, userCommandID, stock, centsInt, numStockInt, true)
		if pushErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified user does not exist")
			break
		}

		if pushErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, pushErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)

	case "POP_USER_SELL":
		userCommandID := data[1]

		sell, popErr := popUserReserve(client, userCommandID, true)
		if popErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The specified user does not exist")
			break
		}

		if popErr == errEmptyStack {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "the sell stack is empty")
			break
		}

		if popErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, popErr.Error())
			break
		}

		sellBytes, jsonErr := json.Marshal(sell)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		lib.ServerSendResponse(conn, lib.StatusOk, string(sellBytes))

	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Data Server Command")
	}
}
