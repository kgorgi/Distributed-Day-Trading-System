package main

import "net"

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

func (client *databaseWrapper) getStocks(userid string) ([]stock, error) {
	return make([]stock, 0), nil
}
