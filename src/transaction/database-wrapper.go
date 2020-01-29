package main

import (
	"net"
)

type databaseWrapper struct {
	client net.Conn
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
	amount = amount + cents
	return nil
}

func (client *databaseWrapper) getBalance(userid string) (uint64, error) {
	return amount, nil
}

func (client *databaseWrapper) removeAmount(userid string, cents uint64) error {
	amount = amount - cents
	return nil
}

func (client *databaseWrapper) getStockAmount(userid string, stockSymbol string) (uint64, error) {
	return stocks[stockSymbol], nil
}

func (client *databaseWrapper) addStock(userid string, stockSymbol string, amount uint64) error {
	stocks[stockSymbol] = stocks[stockSymbol] + amount
	return nil
}

func (client *databaseWrapper) removeStock(userid string, stockSymbol string, amount uint64) error {
	stocks[stockSymbol] = stocks[stockSymbol] - amount
	return nil
}

func (client *databaseWrapper) getStocks(userid string) ([]stock, error) {
	var results = make([]stock, 0)

	for k, v := range stocks {
		var newStock = stock{
			stockSymbol: k,
			numOfStocks: v,
		}
		results = append(results, newStock)
	}
	return results, nil
}
