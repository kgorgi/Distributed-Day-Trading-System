package main

import ( 
    "fmt"
    "net"
    "context"
    "log"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/resolveurl"
    auditclient "extremeWorkload.com/daytrader/lib/audit"
);

var auditClient = auditclient.AuditClient{
	Server: "database",
}

func handleConnection(conn net.Conn, client *mongo.Client) {
    for {
        payload, err := lib.ServerReceiveRequest(conn)
        if err != nil {
            lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
            return
        }

        processCommand(conn, client, payload);
    }
}

func main() {
	fmt.Println("Starting Data server...")

	//hookup to mongo
	clientOptions := options.Client().ApplyURI(resolveurl.DatabaseDBAddress())
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
