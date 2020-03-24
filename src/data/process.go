package main

import (
	"encoding/json"
	"net"

	"extremeWorkload.com/daytrader/lib"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Instead of having a massive switch statement write functions for handling each request
// Better more detailed error handling
// Figure out a nice way to combine UPDATE_TRIGGER_AMOUNT and UPDATE_TRIGGER_PRICE into one command

func generateIsSellBool(isSellString string) bool {
	return isSellString == "true"
}

func processCommand(conn net.Conn, client *mongo.Client, data []string) {
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
		commandBytes := []byte(data[1])

		var updateCommand modelsdata.UpdateUserCommand
		jsonErr := json.Unmarshal(commandBytes, &updateCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
			break
		}

		// If no stock should be added or removed
		if updateCommand.Stock == "" || updateCommand.StockAmount == 0 {
			updateErr := updateCents(client, updateCommand.UserID, updateCommand.Cents)
			if updateErr != nil {
				lib.ServerSendResponse(conn, lib.StatusUserError, "The specified user does not exist, or they do not have the specified amount of money")
			}

			lib.ServerSendOKResponse(conn)
		}

		updateErr := updateStockAndCents(client, updateCommand.UserID, updateCommand.Stock, updateCommand.StockAmount, updateCommand.Cents)
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
		commandBytes := []byte(data[1])

		var readCommand modelsdata.ChooseTriggerCommand
		jsonErr := json.Unmarshal(commandBytes, &readCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		trigger, readErr := readTrigger(client, readCommand.UserID, readCommand.Stock, readCommand.IsSell)
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
		commandBytes := []byte(data[1])

		var deleteCommand modelsdata.ChooseTriggerCommand
		jsonErr := json.Unmarshal(commandBytes, &deleteCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		deletedTrigger, deleteErr := deleteTrigger(client, deleteCommand.UserID, deleteCommand.Stock, deleteCommand.IsSell)
		if deleteErr == errNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, deleteErr.Error())
			break
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
		commandBytes := []byte(data[1])

		var updateCommand modelsdata.UpdateTriggerPriceCommand
		jsonErr := json.Unmarshal(commandBytes, &updateCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		updateErr := updateTriggerPrice(client, updateCommand.UserID, updateCommand.Stock, updateCommand.IsSell, updateCommand.Price)
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
		commandBytes := []byte(data[1])

		var updateCommand modelsdata.UpdateTriggerAmountCommand
		jsonErr := json.Unmarshal(commandBytes, &updateCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		updateErr := updateTriggerAmount(client, updateCommand.UserID, updateCommand.Stock, updateCommand.IsSell, updateCommand.Amount)
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
		commandBytes := []byte(data[1])

		var pushCommand modelsdata.PushUserReserveCommand
		jsonErr := json.Unmarshal(commandBytes, &pushCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		pushErr := pushUserReserve(client, pushCommand.UserID, pushCommand.Stock, pushCommand.Cents, pushCommand.NumStock, false)
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
		commandBytes := []byte(data[1])

		var pushCommand modelsdata.PushUserReserveCommand
		jsonErr := json.Unmarshal(commandBytes, &pushCommand)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		pushErr := pushUserReserve(client, pushCommand.UserID, pushCommand.Stock, pushCommand.Cents, pushCommand.NumStock, true)
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
