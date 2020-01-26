package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func parseCommandRequest(r *http.Request) string {
	vars := mux.Vars(r)
	var command strings.Builder
	command.WriteString(vars["cmd"])
	r.ParseForm()
	for _, v := range r.Form {
		command.WriteString(", " + v[0])
	}
	return command.String()
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
	myRouter.HandleFunc("/command/{cmd}", commandRoute)

	myRouter.HandleFunc("/heartbeat", heartbeat)

	myRouter.Use(loggingMiddleware)

	return myRouter
}

func main() {

	transactionClient := TransactionClient{
		Network:       "tcp",
		RemoteAddress: ":8081",
	}
	transactionClient.ConnectSocket()
	fmt.Println("start server")
	http.ListenAndServe(":9090", getRouter(transactionClient))
}
