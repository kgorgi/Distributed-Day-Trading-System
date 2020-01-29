package main

import "net"

type databaseWrapper struct {
	client net.Conn
}

// IsUserExist check if user is in db
func (client *databaseWrapper) userExists(userid string) (bool, error) {
	return true, nil
}

// CreateUser create user
func (client *databaseWrapper) createUser(userid string) error {
	return nil
}

// AddAmount add money to user balance
func (client *databaseWrapper) addAmount(userid string, cents uint64) error {
	return nil
}
