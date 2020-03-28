package main

import (
	"fmt"
	"strconv"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
)

func buyTrigger(trigger data.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	numOfStocks := trigger.Amount_Cents / stockPrice
	moneyToAdd := trigger.Amount_Cents - (stockPrice * numOfStocks)

	updateErr := data.UpdateUser(trigger.User_Command_ID, trigger.Stock, int(numOfStocks), int(moneyToAdd), auditClient)
	if updateErr != nil {
		return updateErr
	}

	_, deleteErr := data.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	return deleteErr
}

func sellTrigger(trigger data.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	stocksInReserve := trigger.Amount_Cents / trigger.Price_Cents
	moneyToAdd := stockPrice * stocksInReserve

	updateErr := data.UpdateUser(trigger.User_Command_ID, "", 0, int(moneyToAdd), auditClient)
	if updateErr != nil {
		return updateErr
	}

	_, deleteErr := data.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	return deleteErr
}

func checkTriggers(auditClient *auditclient.AuditClient) {
	for {
		lib.Debugln("Checking Triggers")

		triggers, err := data.ReadTriggers()
		for err != nil {
			fmt.Println("Something went wrong, trying again in 10 seconds")
			time.Sleep(10 * time.Second)
			triggers, err = data.ReadTriggers()
		}

		lib.Debugln(strconv.Itoa(len(triggers)) + " Triggers have been fetched, analysing")

		for _, trigger := range triggers {
			auditClient.TransactionNum = trigger.Transaction_Number
			if trigger.Is_Sell {
				auditClient.Command = "SET_SELL_TRIGGER"
			} else {
				auditClient.Command = "SET_BUY_TRIGGER"
			}

			stockPrice, err := GetQuote(trigger.Stock, trigger.User_Command_ID, false, auditClient)
			if err != nil {
				auditClient.LogErrorEvent(err.Error())
				continue
			}

			if trigger.Price_Cents != 0 {
				if trigger.Is_Sell && stockPrice >= trigger.Price_Cents {
					if err := sellTrigger(trigger, stockPrice, auditClient); err != nil {
						auditClient.LogErrorEvent("Sell trigger failed: " + err.Error())
						continue
					}
				} else if !trigger.Is_Sell && stockPrice <= trigger.Price_Cents {
					if err := buyTrigger(trigger, stockPrice, auditClient); err != nil {
						auditClient.LogErrorEvent("Buy trigger failed: " + err.Error())
						continue
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}
