package main

import ( 
    "fmt"
    "net"
    "context"
    "log"
    "strings"
    "encoding/json"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "extremeWorkload.com/daytrader/lib"
    "extremeWorkload.com/daytrader/lib/models/data"
);

func createTrigger(client *mongo.Client, trigger modelsdata.Trigger) error{
    collection := client.Database("extremeworkload").Collection("triggers")
    _, err := collection.InsertOne(context.TODO(), trigger);
    return err
}

func readTriggers(client *mongo.Client) ([]modelsdata.Trigger, error) {
    collection := client.Database("extremeworkload").Collection("triggers")
    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        return nil, err
    }

    //copy over users
    var triggers []modelsdata.Trigger
    defer cursor.Close(context.TODO())
    for cursor.Next(context.TODO()) {
        var trigger modelsdata.Trigger
        cursor.Decode(&trigger)
        triggers = append(triggers, trigger);
    }

    return triggers, nil
}

func readTrigger(client *mongo.Client, user_command_ID string, stock string) (modelsdata.Trigger, error) {
    collection := client.Database("extremeworkload").Collection("triggers")

    var trigger modelsdata.Trigger
    err := collection.FindOne(context.TODO(), bson.M{"user_command_id": user_command_ID, "stock": stock}).Decode(&trigger);
    return trigger, err
}

func updateTrigger(client *mongo.Client, trigger modelsdata.Trigger) error {
    collection := client.Database("extremeworkload").Collection("triggers")
    update := bson.D{
        {"$set", bson.M{"price_cents": trigger.Price_Cents, "amount_cents": trigger.Amount_Cents}},
    }
    filter := bson.M{"user_command_id": trigger.User_Command_ID, "stock": trigger.Stock}
    _, err := collection.UpdateOne(context.TODO(), filter, update)
    return err
}

func deleteTrigger(client *mongo.Client, user_command_ID string, stock string) error {
    collection := client.Database("extremeworkload").Collection("triggers")
    filter := bson.M{"user_command_id": user_command_ID, "stock": stock};
    _, err := collection.DeleteOne(context.TODO(), filter);
    return err
}

func createUser(client *mongo.Client, user  modelsdata.User) error{
    collection := client.Database("extremeworkload").Collection("users")
    _, err := collection.InsertOne(context.TODO(), user);
    return err
}

func readUsers(client *mongo.Client) ([]modelsdata.User, error) {
    collection := client.Database("extremeworkload").Collection("users")
    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        return nil, err
    }

    //copy over users
    var users []modelsdata.User
    defer cursor.Close(context.TODO())
    for cursor.Next(context.TODO()) {
        var user modelsdata.User
        cursor.Decode(&user)
        users = append(users, user);
    }

    return users, nil
}

func readUser(client *mongo.Client, command_ID string) (modelsdata.User, error) {
    collection := client.Database("extremeworkload").Collection("users")

    var user modelsdata.User
    err := collection.FindOne(context.TODO(), bson.D{{"command_id", command_ID}}).Decode(&user);
    return user, err
}

func updateUser(client *mongo.Client, user modelsdata.User) error {
    collection := client.Database("extremeworkload").Collection("users")

    update := bson.D{
        {"$set", bson.M{"cents": user.Cents, "investments": user.Investments}},
    }

    filter := bson.M{"command_id": user.Command_ID}

    _, err := collection.UpdateOne(context.TODO(), filter, update)
    return err
}

func deleteUser(client *mongo.Client, command_ID string) error {
    collection := client.Database("extremeworkload").Collection("users")
    filter := bson.M{"command_id": command_ID};
    _, err := collection.DeleteOne(context.TODO(), filter);
    return err
}

func handleConnection(conn net.Conn, client *mongo.Client) {
    for {
        payload, err := lib.ServerReceiveRequest(conn)
        if err != nil {
            lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
            return
        }
        data := strings.Split(payload, "|")
        switch data[0] {
        case "CREATE_USER":
            userJson := data[1]
            var newUser modelsdata.User
            jsonErr := json.Unmarshal([]byte(userJson), &newUser)
            if jsonErr != nil {
                lib.ServerSendResponse(conn, 400, "json input is incorrect");
                break;
            }
            
            createErr := createUser(client, newUser)
            if createErr != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "everythings good my dude")
        case "READ_USER":
            commandID := data[1];
            user, readError := readUser(client, commandID)
			userBytes, jsonError := json.Marshal(user)

			if(readError != nil && jsonError != nil) {
				lib.ServerSendResponse(conn, 500, "something went wrong");
			}

			lib.ServerSendResponse(conn, 200, string(userBytes));
        case "READ_USERS":
            users, readError := readUsers(client)
            usersBytes, jsonError := json.Marshal(users)
			
            if readError != nil || jsonError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }
			
            lib.ServerSendResponse(conn, 200, string(usersBytes))
        case "UPDATE_USER":
            userJson := data[1]
            var userUpdate modelsdata.User
            jsonError := json.Unmarshal([]byte(userJson), &userUpdate)
            if jsonError != nil {
                lib.ServerSendResponse(conn, 400, "json input is incorrect");
                break;
            }
            
            updateError := updateUser(client, userUpdate)
            if updateError != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "everythings good my dude")
        case "DELETE_USER":
            commandID := data[1];
            deleteError := deleteUser(client, commandID)
            
            if deleteError != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "user has been deleted");
        case "CREATE_TRIGGER":
            triggerJson := data[1]
            var newTrigger modelsdata.Trigger
            jsonError := json.Unmarshal([]byte(triggerJson), &newTrigger)
            if jsonError != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            createErr := createTrigger(client, newTrigger)
            if createErr != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "trigger created!")
        case "READ_TRIGGER":
            userCommandID := data[1];
            stock := data[2];

            trigger, readError := readTrigger(client, userCommandID, stock);
            triggerBytes, jsonError := json.Marshal(trigger);

            if(readError != nil && jsonError != nil) {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }
            
            lib.ServerSendResponse(conn, 200, string(triggerBytes));
        case "READ_TRIGGERS":
            triggers, readError := readTriggers(client);
            triggersBytes, jsonError := json.Marshal(triggers)

            if readError != nil || jsonError != nil {
				lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, string(triggersBytes));
        case "UPDATE_TRIGGER":
            triggerJson := data[1];
            var triggerUpdate modelsdata.Trigger
            jsonError := json.Unmarshal([]byte(triggerJson), &triggerUpdate);

            if jsonError != nil {
                lib.ServerSendResponse(conn, 400, "json input is incorrect");
                break;
            }

            updateError := updateTrigger(client, triggerUpdate)
            if updateError != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "trigger updated");
        case "DELETE_TRIGGER":
            userCommandID := data[1];
            stock := data[2];

            deleteError := deleteTrigger(client, userCommandID, stock);
            if deleteError != nil {
                lib.ServerSendResponse(conn, 500, "something went wrong");
                break;
            }

            lib.ServerSendResponse(conn, 200, "trigger has been deleted");


        default: lib.ServerSendResponse(conn, 400, "Invalid Data Server Command")

        }
        lib.ServerSendOKResponse(conn)
    }
}

func main() {
    fmt.Println("Starting Data server...")

    //hookup to mongo
    clientOptions := options.Client().ApplyURI("mongodb://data-mongoDB:27017/mongodb")
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB!")
    
    //start listening on the port
    ln, err := net.Listen("tcp", ":5001")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("Started Server on Port 5001")

    //connection handling
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        fmt.Println("Connection Established")
        go handleConnection(conn, client)
    }
}

