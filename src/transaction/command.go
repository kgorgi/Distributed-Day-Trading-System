package main

import "encoding/json"

// Command struct for command
type Command struct {
	Command     string
	Userid      string
	Amount      string
	StockSymbol string
}

// CommandFromJSON creates a command from a json string
func CommandFromJSON(jsonCommand string, transactionCommand *Command) error {
	var err error
	err = json.Unmarshal([]byte(jsonCommand), &transactionCommand)
	if err != nil {
		return err
	}
	return nil
}

func (command *Command) isUseridValid() (bool, error) {
	if true {
		return true, nil
	}
	return false, &InvalidData{command.Userid + " is not a valid userid"}
}

func (command *Command) isAmountValid() (bool, error) {
	if true {
		return true, nil
	}
	return false, &InvalidData{command.Amount + " is not a valid amount"}
}
