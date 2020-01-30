package main

import (
	"net"
	"sync"
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

// AddAmount add money to user balance
func (client *databaseWrapper) addAmount(userid string, cents uint64) error {
	client.mux.Lock()
	amount = amount + cents
	client.mux.Unlock()

	return nil
}

func (client *databaseWrapper) getBalance(userid string) (uint64, error) {
	client.mux.Lock()
	cents := amount
	client.mux.Unlock()

	return cents, nil
}

func (client *databaseWrapper) removeAmount(userid string, cents uint64) error {
	client.mux.Lock()
	amount = amount - cents
	client.mux.Unlock()

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
