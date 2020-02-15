package main

import (
	"fmt"
	"time"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func buyTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	numOfStocks := trigger.Amount_Cents / stockPrice
	moneyToAdd := trigger.Amount_Cents - (stockPrice * numOfStocks)

	user, readErr := dataclient.ReadUser(trigger.User_Command_ID)
	if readErr != nil {
		return readErr
	}
	user.Cents += moneyToAdd
	user.Investments = addStock(user.Investments, trigger.Stock, numOfStocks)

	updateErr := dataclient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       user.Command_ID,
		FundsInCents: moneyToAdd,
	})

	deleteErr := dataclient.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func sellTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	stocksInReserve := trigger.Amount_Cents / trigger.Price_Cents
	moneyToAdd := stockPrice * stocksInReserve

	user, readErr := dataclient.ReadUser(trigger.User_Command_ID)
	if readErr != nil {
		return readErr
	}
	user.Cents += moneyToAdd

	updateErr := dataclient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       user.Command_ID,
		FundsInCents: moneyToAdd,
	})

	deleteErr := dataclient.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func checkTriggers(auditClient auditclient.AuditClient) {
	for {
		fmt.Println("Checking Triggers")

		triggers, err := dataclient.ReadTriggers()
		for err != nil {
			fmt.Println("Something went wrong, trying again in 10 seconds")
			time.Sleep(10 * time.Second)
			triggers, err = dataclient.ReadTriggers()
		}

		fmt.Println(string(len(triggers)) + " Triggers have been fetched, analysing")

		for _, trigger := range triggers {
			stockPrice := GetQuote(trigger.Stock, trigger.User_Command_ID, &auditClient)
			if trigger.Price_Cents != 0 {
				if trigger.Is_Sell && stockPrice >= trigger.Price_Cents {
					if err := sellTrigger(trigger, stockPrice, &auditClient); err != nil {
						fmt.Println(err)
						continue
					}
				} else if !trigger.Is_Sell && stockPrice <= trigger.Price_Cents {
					if err := buyTrigger(trigger, stockPrice, &auditClient); err != nil {
						fmt.Println(err)
						continue
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}
