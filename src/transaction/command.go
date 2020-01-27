package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var isAlphanumeric = regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
var isNumeric = regexp.MustCompile(`^[0-9]+$`).MatchString

// Command struct for command
type Command struct {
	Command     string
	Userid      string
	Amount      string
	Cents       int
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
	if isAlphanumeric(command.Userid) {
		return true, nil
	}
	return false, &InvalidData{command.Userid + " is not a valid userid"}
}

// GetCents parses Amount into cents and updates Cents field
func (command *Command) GetCents() (int, error) {
	amountStrings := strings.Split(command.Amount, ".")
	cents := 0
	for i, v := range amountStrings {
		val, err := strconv.Atoi(v)
		if err != nil || val < 0 || i > 1 {
			return -1, &InvalidData{command.Amount + " is not a valid amount"}
		}
		cents += (100 ^ (1 - i)) * val
	}
	command.Cents = cents
	return cents, nil
}
