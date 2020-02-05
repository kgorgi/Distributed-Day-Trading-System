package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"strconv"
	"github.com/gorilla/mux"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

const webServerAddress = ":8080"
var transactionNum uint64 = 0

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
func commandRoute(w http.ResponseWriter, r *http.Request, ) {
	command := parseCommandRequest(r)

	var nextNum = atomic.AddUint64(&transactionNum, 1)
	command["transactionNum"] = strconv.FormatUint(nextNum, 10)

	var auditClient = auditclient.AuditClient{
		Command: command["command"],
		Server: "web",
		TransactionNum: nextNum,
	}

	auditInfo := auditclient.UserCommandInfo{
		OptionalUserID: command["userid"],
		OptionalFilename: command["filename"],  
		OptionalStockSymbol:  command["stockSymbol"], 
	}

	if command["amount"] != "" {
		funds, _ := strconv.ParseUint(command["amount"], 10, 64)
		auditInfo.OptionalFundsInCents = &funds
	}

	auditClient.LogUserCommandRequest(auditInfo)

	var message string
	var err error
	var status int
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
		fmt.Println(r.RequestURI)
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
	fmt.Println("start server")
	http.ListenAndServe(webServerAddress, getRouter())
}
