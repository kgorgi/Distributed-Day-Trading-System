package main

import (
	"fmt"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func buyTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	numOfStocks := trigger.Amount_Cents / stockPrice
	moneyToAdd := trigger.Amount_Cents - (stockPrice * numOfStocks)

	updateErr := dataclient.UpdateUser(trigger.User_Command_ID, trigger.Stock, int(numOfStocks), int(moneyToAdd), auditClient)
	if updateErr != nil {
		return updateErr
	}

	_, deleteErr := dataclient.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell, auditClient)
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func sellTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	stocksInReserve := trigger.Amount_Cents / trigger.Price_Cents
	moneyToAdd := stockPrice * stocksInReserve

	updateErr := dataclient.UpdateUser(trigger.User_Command_ID, "", 0, int(moneyToAdd), auditClient)
	if updateErr != nil {
		return updateErr
	}

	_, deleteErr := dataclient.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell, auditClient)
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func checkTriggers(auditClient *auditclient.AuditClient) {
	for {
		lib.Debugln("Checking Triggers")

		triggers, err := dataclient.ReadTriggers(auditClient)
		for err != nil {
			fmt.Println("Something went wrong, trying again in 10 seconds")
			time.Sleep(10 * time.Second)
			triggers, err = dataclient.ReadTriggers(auditClient)
		}

		lib.Debugln(string(len(triggers)) + " Triggers have been fetched, analysing")

		for _, trigger := range triggers {
			auditClient.TransactionNum = trigger.Transaction_Number
			if trigger.Is_Sell {
				auditClient.Command = "SET_SELL_TRIGGER"
			} else {
				auditClient.Command = "SET_BUY_TRIGGER"
			}

			stockPrice := GetQuote(trigger.Stock, trigger.User_Command_ID, false, auditClient)
			if trigger.Price_Cents != 0 {
				if trigger.Is_Sell && stockPrice >= trigger.Price_Cents {
					if err := sellTrigger(trigger, stockPrice, auditClient); err != nil {
						fmt.Println(err)
						continue
					}
				} else if !trigger.Is_Sell && stockPrice <= trigger.Price_Cents {
					if err := buyTrigger(trigger, stockPrice, auditClient); err != nil {
						fmt.Println(err)
						continue
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}
