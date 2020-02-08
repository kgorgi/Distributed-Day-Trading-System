package main

import (
	"encoding/json"
	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/mongo"
	"net"
	"strings"
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
		userJson := data[1]
		var userUpdate modelsdata.User
		jsonErr := json.Unmarshal([]byte(userJson), &userUpdate)
		if jsonErr != nil {
			lib.ServerSendResponse(conn, lib.StatusUserError, jsonErr.Error())
			break
		}

		updateErr := updateUser(client, userUpdate)
		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)
	case "DELETE_USER":
		commandID := data[1]
		deleteErr := deleteUser(client, commandID)

		if deleteErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
			break
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

		deleteErr := deleteTrigger(client, userCommandID, stock, generateIsSellBool(isSellString))
		if deleteErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
			break
		}

		lib.ServerSendOKResponse(conn)
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid Data Server Command")
	}
}
