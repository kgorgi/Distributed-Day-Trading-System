package main

import (
	"fmt"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func buyTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	if stockPrice > trigger.Amount_Cents {
		return nil
	}

	numOfStocks := trigger.Amount_Cents / stockPrice
	moneyToAdd := trigger.Amount_Cents - (stockPrice * numOfStocks)

	err := dataConn.addAmount(trigger.User_Command_ID, moneyToAdd, auditClient)
	if err != nil {
		return err
	}

	err = dataConn.addStock(trigger.User_Command_ID, trigger.Stock, numOfStocks)
	if err != nil {
		return err
	}

	err = dataConn.deleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	if err != nil {
		return err
	}

	return nil
}

func sellTrigger(trigger modelsdata.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	if stockPrice > trigger.Amount_Cents {
		return nil
	}

	stocksInReserve := trigger.Amount_Cents / trigger.Price_Cents
	numOfStocksToSell := trigger.Amount_Cents / stockPrice
	if numOfStocksToSell == 0 {
		numOfStocksToSell = 1
	}

	moneyToAdd := stockPrice * numOfStocksToSell
	stocksRemaining := stocksInReserve - numOfStocksToSell

	err := dataConn.addAmount(trigger.User_Command_ID, moneyToAdd, auditClient)
	if err != nil {
		return err
	}

	if stocksRemaining > 0 {
		err = dataConn.addStock(trigger.User_Command_ID, trigger.Stock, stocksRemaining)
		if err != nil {
			return err
		}
	}

	err = dataConn.deleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell)
	if err != nil {
		return err
	}

	return nil
}

func checkTriggers(auditClient auditclient.AuditClient) {
	for {
		lib.Debugln("Checking Triggers")

		triggers, err := dataConn.getTriggers()
		for err != nil {
			fmt.Println("Something went wrong, trying again in 10 seconds")
			time.Sleep(10 * time.Second)
			triggers, err = dataConn.getTriggers()
		}

		lib.Debugln(string(len(triggers)) + " Triggers have been fetched, analysing")

		for _, trigger := range triggers {
			stockPrice := GetQuote(trigger.Stock, trigger.User_Command_ID, &auditClient)
			if trigger.Price_Cents != 0 && stockPrice >= trigger.Price_Cents {
				if trigger.Is_Sell {
					if err := sellTrigger(trigger, stockPrice, &auditClient); err != nil {
						fmt.Println(err)
						continue
					}
				} else {
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
