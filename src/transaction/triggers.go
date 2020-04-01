package main

import (
	"context"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
)

var isTriggerCheckingActive bool

func buyTrigger(trigger data.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	numOfStocks := trigger.Amount_Cents / stockPrice
	moneyToAdd := trigger.Amount_Cents - (stockPrice * numOfStocks)

	err, _ := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		updateErr := data.UpdateUser(trigger.User_Command_ID, trigger.Stock, int(numOfStocks), int(moneyToAdd), ctx, auditClient)
		if updateErr != nil {
			return updateErr, lib.StatusSystemError
		}

		_, deleteErr := data.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell, ctx)
		return deleteErr, lib.StatusSystemError
	})

	return err
}

func sellTrigger(trigger data.Trigger, stockPrice uint64, auditClient *auditclient.AuditClient) error {
	stocksInReserve := trigger.Amount_Cents / trigger.Price_Cents
	moneyToAdd := stockPrice * stocksInReserve

	err, _ := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		updateErr := data.UpdateUser(trigger.User_Command_ID, "", 0, int(moneyToAdd), ctx, auditClient)
		if updateErr != nil {
			return updateErr, lib.StatusSystemError
		}

		_, deleteErr := data.DeleteTrigger(trigger.User_Command_ID, trigger.Stock, trigger.Is_Sell, ctx)
		return deleteErr, lib.StatusSystemError
	})

	return err
}

func checkTriggers(auditClient *auditclient.AuditClient) {
	for {
		if !isTriggerCheckingActive {
			lib.Debugln("Trigger Checking is not active on this sever")
			time.Sleep(15 * time.Second)
		}
		lib.Debugln("Checking Triggers")

		triggerIterator, readErr := data.CheckTriggersIterator()
		for readErr != nil {
			lib.Errorln("Something went wrong, trying again in 10 seconds " + readErr.Error())
			time.Sleep(10 * time.Second)
			triggerIterator, readErr = data.CheckTriggersIterator()
		}

		var triggersChecked = 0
		for {
			validTrigger, trigger, err := triggerIterator()

			if err != nil {
				lib.Errorln("Failed to check trigger " + err.Error())
				continue
			}

			if !validTrigger {
				break
			}

			stockPrice, err := GetQuote(trigger.Stock, trigger.User_Command_ID, false, auditClient)
			if err != nil {
				auditClient.LogErrorEvent(err.Error())
				continue
			}

			auditClient.TransactionNum = trigger.Transaction_Number
			if trigger.Is_Sell {
				auditClient.Command = "SET_SELL_TRIGGER"
			} else {
				auditClient.Command = "SET_BUY_TRIGGER"
			}

			if trigger.Is_Sell && stockPrice >= trigger.Price_Cents {
				err := sellTrigger(trigger, stockPrice, auditClient)
				if err != nil {
					auditClient.LogErrorEvent("Sell trigger failed: " + err.Error())
				}
			} else if !trigger.Is_Sell && stockPrice <= trigger.Price_Cents {
				err := buyTrigger(trigger, stockPrice, auditClient)
				if err != nil {
					auditClient.LogErrorEvent("Buy trigger failed: " + err.Error())
				}
			}

			triggersChecked++
		}

		lib.Debug("Checked %d triggers.\n", triggersChecked)

		time.Sleep(60 * time.Second)
	}
}
