package main

import ( 
    "net"
    "strings"
    "encoding/json"
    "extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/models/data"
	"go.mongodb.org/mongo-driver/mongo"
);

func generateIsSellBool(isSellString string) bool {
    if isSellString == "true" {
        return true
    }else{
        return false
    }
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
				lib.ServerSendResponse(conn, 400, jsonErr.Error());
				break;
			}
			
			createErr := createUser(client, newUser)
			if createErr != nil {
				lib.ServerSendResponse(conn, 500, createErr.Error());
				break;
			}

			lib.ServerSendOKResponse(conn)
        case "READ_USER":
            commandID := data[1];
            user, readErr := readUser(client, commandID)
            if readErr != nil {
                status := 500
                if readErr == mongo.ErrNoDocuments {
                    status = 404
                }
                lib.ServerSendResponse(conn, status, readErr.Error())
                break;
            }

            userBytes, jsonErr := json.Marshal(user)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }
            
			lib.ServerSendResponse(conn, lib.StatusOk, string(userBytes));
        case "READ_USERS":
            users, readErr := readUsers(client)
            if readErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }

            usersBytes, jsonErr := json.Marshal(users)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }
			
            lib.ServerSendResponse(conn, lib.StatusOk, string(usersBytes))
        case "UPDATE_USER":
            userJson := data[1]
            var userUpdate modelsdata.User
            jsonErr := json.Unmarshal([]byte(userJson), &userUpdate)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 400, jsonErr.Error());
                break;
            }
            
            updateErr := updateUser(client, userUpdate)
            if updateErr != nil {
                lib.ServerSendResponse(conn, 500, updateErr.Error());
                break;
            }

            lib.ServerSendOKResponse(conn)
        case "DELETE_USER":
            commandID := data[1];
            deleteErr := deleteUser(client, commandID)
            
            if deleteErr != nil {
                lib.ServerSendResponse(conn, 500, deleteErr.Error());
                break;
            }

            lib.ServerSendOKResponse(conn)
        case "CREATE_TRIGGER":
            triggerJson := data[1]
            var newTrigger modelsdata.Trigger
            jsonErr := json.Unmarshal([]byte(triggerJson), &newTrigger)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 500, jsonErr.Error());
                break;
            }

            createErr := createTrigger(client, newTrigger)
            if createErr != nil {
                lib.ServerSendResponse(conn, 500, createErr.Error());
                break;
            }

            lib.ServerSendOKResponse(conn)
        case "READ_TRIGGER":
            userCommandID := data[1];
            stock := data[2];
            isSellString := data[3];

            trigger, readErr := readTrigger(client, userCommandID, stock, generateIsSellBool(isSellString))
            if readErr != nil {
                status := 500
                if readErr == mongo.ErrNoDocuments {
                    status = 404
                }
                lib.ServerSendResponse(conn, status, readErr.Error())
                break;
            }

            triggerBytes, jsonErr := json.Marshal(trigger);
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }

            lib.ServerSendResponse(conn, lib.StatusOk, string(triggerBytes));
        case "READ_TRIGGERS":
            triggers, readErr := readTriggers(client);
            if readErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }

            triggersBytes, jsonErr := json.Marshal(triggers)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 500, readErr.Error())
                break;
            }

            lib.ServerSendResponse(conn, lib.StatusOk, string(triggersBytes));
        case "UPDATE_TRIGGER":
            triggerJson := data[1];
            var triggerUpdate modelsdata.Trigger
            jsonErr := json.Unmarshal([]byte(triggerJson), &triggerUpdate);

            if jsonErr != nil {
                lib.ServerSendResponse(conn, 400, jsonErr.Error());
                break;
            }

            updateErr := updateTrigger(client, triggerUpdate)
            if updateErr != nil {
                lib.ServerSendResponse(conn, 500, updateErr.Error());
                break;
            }

            lib.ServerSendOKResponse(conn)
        case "DELETE_TRIGGER":
            userCommandID := data[1];
            stock := data[2];
            isSellString := data[3];

            deleteErr := deleteTrigger(client, userCommandID, stock, generateIsSellBool(isSellString));
            if deleteErr != nil {
                lib.ServerSendResponse(conn, 500, deleteErr.Error());
                break;
            }

            lib.ServerSendOKResponse(conn)
        default: lib.ServerSendResponse(conn, 400, "Invalid Data Server Command")
	}
}