package main

import (
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
)

var isTriggerCheckingActive bool

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
	var currentInstanceNum uint64 = 1
	for {
		if !isTriggerCheckingActive {
			lib.Debugln("Trigger Checking is not active on this sever")
		} else {
			go checkAllTriggers(currentInstanceNum, *auditClient)
			currentInstanceNum++
		}

		time.Sleep(60 * time.Second)
	}
}

func checkAllTriggers(instanceNum uint64, auditClient auditclient.AuditClient) {
	lib.Debug("Triggers Check #%d. Checking Triggers.\n", instanceNum)

	triggerIterator, readErr := data.CheckTriggersIterator()
	var currentAttempt = 0
	for readErr != nil {
		lib.Error("Triggers Check #%d: Something went wrong, trying again in 10 seconds %s\n", instanceNum, readErr.Error())
		time.Sleep(5 * time.Second)

		currentAttempt++
		if currentAttempt > 3 {
			lib.Error("Triggers Check #%d. Exhausted all attempts to read triggers from DB\n", instanceNum)
			return
		}
		triggerIterator, readErr = data.CheckTriggersIterator()
	}

	var triggersChecked = 0
	for {
		validTrigger, trigger, err := triggerIterator()

		if err != nil {
			lib.Error("Triggers Check #%d. Failed to check trigger %s\n", instanceNum, err.Error())
			continue
		}

		if !validTrigger {
			break
		}

		stockPrice, err := GetQuote(trigger.Stock, trigger.User_Command_ID, false, &auditClient)
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
			err := sellTrigger(trigger, stockPrice, &auditClient)
			if err != nil {
				auditClient.LogErrorEvent("Sell trigger failed: " + err.Error())
			}
		} else if !trigger.Is_Sell && stockPrice <= trigger.Price_Cents {
			err := buyTrigger(trigger, stockPrice, &auditClient)
			if err != nil {
				auditClient.LogErrorEvent("Buy trigger failed: " + err.Error())
			}
		}

		triggersChecked++
	}

	lib.Debug("Triggers Check #%d. Checked %d triggers.\n", instanceNum, triggersChecked)
}
