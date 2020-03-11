package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

const webServerAddress = ":8080"

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
func commandRoute(w http.ResponseWriter, r *http.Request) {
	command := parseCommandRequest(r)

	var message string
	var err error
	var status int

	isValid, status, message := validateParameters(command)
	if !isValid {
		w.WriteHeader(status)
		w.Write([]byte(message))
		return
	}

	var auditClient = auditclient.AuditClient{
		Server:         "web",
		TransactionNum: 0,
	}

	auditInfo := auditclient.UserCommandInfo{
		OptionalUserID:      command["userid"],
		OptionalFilename:    command["filename"],
		OptionalStockSymbol: command["stockSymbol"],
	}

	if command["amount"] != "" {
		funds, _ := strconv.ParseUint(command["amount"], 10, 64)
		auditInfo.OptionalFundsInCents = &funds
	}

	transactionNum := auditClient.LogUserCommandRequest(auditInfo)
	command["transactionNum"] = strconv.FormatUint(transactionNum, 10)

	if command["command"] == "DUMPLOG" {
		message, err = auditClient.DumpLogAll()
		status = 200
	} else {
		var transactionClient TransactionClient
		status, message, err = transactionClient.SendCommand(command)
	}

	w.WriteHeader(status)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(message))
	}
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lib.Debugln(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func getRouter() http.Handler {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/command/{command_name}", commandRoute)

	myRouter.HandleFunc("/heartbeat", heartbeat)

	myRouter.Use(loggingMiddleware)

	return myRouter
}

func main() {
	initParameterMaps()

	fmt.Println("Starting web server...")
	err := http.ListenAndServeTLS(webServerAddress, "./ssl/cert.pem", "./ssl/key.pem", getRouter())
	if err != nil {
		fmt.Println(err.Error())
	}
}
