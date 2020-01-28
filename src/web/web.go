package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

var auditClient = auditclient.AuditClient{
	Server: "web",
}

func parseCommandRequest(r *http.Request) map[string]string {

	command := make(map[string]string)

	vars := mux.Vars(r)

	command["command"] = vars["command_name"]
	r.ParseForm()
	for k, v := range r.Form {
		command[k] = v[0]
	}
	return command
}

// Creates a route method. Whenever the route is called, it always uses the same socket
func createCommandRoute(transactionClient TransactionClient) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		command := parseCommandRequest(r)
		status, message, err := transactionClient.sendCommand(command)

		w.WriteHeader(status)
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(message))
		}

	}

}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func getRouter(transactionClient TransactionClient) http.Handler {
	myRouter := mux.NewRouter().StrictSlash(true)

	commandRoute := createCommandRoute(transactionClient)
	myRouter.HandleFunc("/command/{command_name}", commandRoute)

	myRouter.HandleFunc("/heartbeat", heartbeat)

	myRouter.Use(loggingMiddleware)

	return myRouter
}

func main() {

	auditClient.LogDebugEvent(auditclient.DebugEventInfo{
		TransactionNum:       -1,
		Command:              "N/A",
		OptionalDebugMessage: "Starting Web Server",
	})

	transactionClient := TransactionClient{
		Network:       "tcp",
		RemoteAddress: "transaction-server:5000",
		// RemoteAddress: ":5000",
	}
	transactionClient.ConnectSocket()
	fmt.Println("start server")
	http.ListenAndServe(":8080", getRouter(transactionClient))
}
