package main

import (
	"net"
	"strconv"
	"sync"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

type databaseWrapper struct {
	client net.Conn
	mux    sync.Mutex
}

type stock struct {
	stockSymbol string
	numOfStocks uint64
}

// IsUserExist check if user is in db
func (client *databaseWrapper) userExists(userid string) (bool, error) {
	return true, nil
}

// CreateUser create user
func (client *databaseWrapper) createUser(userid string) error {
	return nil
}

var amount = uint64(0)
var stocks = make(map[string]uint64)
var triggers = make(map[string]*modelsdata.Trigger)

// AddAmount add money to user balance
func (client *databaseWrapper) addAmount(
	userid string,
	cents uint64,
	auditClient *auditclient.AuditClient) error {
	client.mux.Lock()
	amount = amount + cents
	client.mux.Unlock()

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       userid,
		FundsInCents: cents,
	})

	return nil
}

func (client *databaseWrapper) getBalance(userid string) (uint64, error) {
	client.mux.Lock()
	cents := amount
	client.mux.Unlock()

	return cents, nil
}

func (client *databaseWrapper) removeAmount(userid string,
	cents uint64,
	auditClient *auditclient.AuditClient) error {
	client.mux.Lock()
	amount = amount - cents
	client.mux.Unlock()

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "remove",
		UserID:       userid,
		FundsInCents: cents,
	})

	return nil
}

func (client *databaseWrapper) getStockAmount(userid string, stockSymbol string) (uint64, error) {
	client.mux.Lock()
	numOfStocks := stocks[stockSymbol]
	client.mux.Unlock()
	return numOfStocks, nil
}

func (client *databaseWrapper) addStock(userid string, stockSymbol string, amount uint64) error {
	client.mux.Lock()
	stocks[stockSymbol] = stocks[stockSymbol] + amount
	client.mux.Unlock()
	return nil
}

func (client *databaseWrapper) removeStock(userid string, stockSymbol string, amount uint64) error {
	client.mux.Lock()
	stocks[stockSymbol] = stocks[stockSymbol] - amount
	client.mux.Unlock()
	return nil
}

func (client *databaseWrapper) getStocks(userid string) ([]stock, error) {
	client.mux.Lock()

	var results = make([]stock, 0)

	for k, v := range stocks {
		var newStock = stock{
			stockSymbol: k,
			numOfStocks: v,
		}
		results = append(results, newStock)
	}
	client.mux.Unlock()
	return results, nil
}

func (client *databaseWrapper) getTrigger(userid string, stockSymbol string, isSell bool) (*modelsdata.Trigger, error) {
	client.mux.Lock()
	var trigger = triggers[triggerKey(userid, stockSymbol, isSell)]
	client.mux.Unlock()

	return trigger, nil
}

func (client *databaseWrapper) createTrigger(userid string, stockSymbol string, cents uint64, isSell bool) error {
	var newTrigger = new(modelsdata.Trigger)
	newTrigger.User_Command_ID = userid
	newTrigger.Stock = stockSymbol
	newTrigger.Amount_Cents = cents
	newTrigger.Is_Sell = isSell

	client.mux.Lock()
	triggers[triggerKey(userid, stockSymbol, isSell)] = newTrigger
	client.mux.Unlock()
	return nil
}

func (client *databaseWrapper) setTriggerAmount(userid string, stockSymbol string, cents uint64, isSell bool) error {
	client.mux.Lock()
	triggers[triggerKey(userid, stockSymbol, isSell)].Price_Cents = cents
	client.mux.Unlock()
	return nil
}

func (client *databaseWrapper) deleteTrigger(userid string, stockSymbol string, isSell bool) error {
	client.mux.Lock()
	delete(triggers, triggerKey(userid, stockSymbol, isSell))
	client.mux.Unlock()
	return nil
}

func (client *databaseWrapper) getTriggers() ([]modelsdata.Trigger, error) {
	results := make([]modelsdata.Trigger, len(triggers))

	client.mux.Lock()

	i := 0
	for _, value := range triggers {
		results[i] = *value
		i++
	}

	client.mux.Unlock()

	return results, nil
}

func triggerKey(userID string, stockSymbol string, isSell bool) string {
	var key = userID + ": " + stockSymbol + ":" + strconv.FormatBool(isSell)
	return key
}
