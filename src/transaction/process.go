package main

import (
	"encoding/json"
	"errors"
	"net"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func processCommand(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	valid := validateParameters(conn, jsonCommand)
	if !valid {
		return
	}

	// Process Command
	switch jsonCommand.Command {
	case "ADD":
		handleAdd(conn, jsonCommand, &auditClient)
	case "QUOTE":
		handleQuote(conn, jsonCommand, &auditClient)
	case "BUY":
		handleBuy(conn, jsonCommand, &auditClient)
	case "COMMIT_BUY":
		handleCommitBuy(conn, jsonCommand, &auditClient)
	case "CANCEL_BUY":
		handleCancelBuy(conn, jsonCommand, &auditClient)
	case "SELL":
		handleSell(conn, jsonCommand, &auditClient)
	case "COMMIT_SELL":
		handleCommitSell(conn, jsonCommand, &auditClient)
	case "CANCEL_SELL":
		handleCancelSell(conn, jsonCommand, &auditClient)
	case "SET_BUY_AMOUNT":
		handleSetBuyAmount(conn, jsonCommand, &auditClient)
	case "SET_BUY_TRIGGER":
		handleSetBuyTrigger(conn, jsonCommand, &auditClient)
	case "CANCEL_SET_BUY":
		handleCancelSetBuy(conn, jsonCommand, &auditClient)
	case "SET_SELL_AMOUNT":
		handleSetSellAmount(conn, jsonCommand, &auditClient)
	case "SET_SELL_TRIGGER":
		handleSetSellTrigger(conn, jsonCommand, &auditClient)
	case "CANCEL_SET_SELL":
		handleCancelSetSell(conn, jsonCommand, &auditClient)
	case "DISPLAY_SUMMARY":
		handleDisplaySummary(conn, jsonCommand, auditClient)
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid command")
	}

}

func addStock(investments []modelsdata.Investment, stockSymbol string, amount uint64) []modelsdata.Investment {
	// Find the investment in the user struct and set the amount specified in the params
	investmentIndex := -1
	for i, investment := range investments {
		if investment.Stock == stockSymbol {
			investmentIndex = i
		}
	}

	// If you can't find the investment create a new investment, otherwise add to the existing one
	if investmentIndex == -1 {
		return append(investments, modelsdata.Investment{stockSymbol, amount})
	} else {
		investments[investmentIndex].Amount += amount
		return investments
	}
}

func findStockAmount(investments []modelsdata.Investment, stockSymbol string) uint64 {
	for _, investment := range investments {
		if investment.Stock == stockSymbol {
			return investment.Amount
		}
	}
	return 0
}

func removeStock(investments []modelsdata.Investment, stockSymbol string, amount uint64) ([]modelsdata.Investment, error) {
	investmentIndex := -1
	for i, investment := range investments {
		if investment.Stock == stockSymbol {
			investmentIndex = i
		}
	}

	// Make sure the stock is found
	if investmentIndex == -1 {
		return []modelsdata.Investment{}, errors.New("this user does not have any of the stock " + stockSymbol)
	}

	// Make sure they have enough stock to remove the amount
	stockAmount := investments[investmentIndex].Amount
	if stockAmount < amount {
		return []modelsdata.Investment{}, errors.New("The user does not have sufficient stock ( " + string(stockAmount) + " ) to remove " + string(amount))
	}

	investments[investmentIndex].Amount -= amount

	// If the remaining amount is 0 remove the investment from the user
	if investments[investmentIndex].Amount == 0 {
		investments[investmentIndex] = investments[len(investments)-1]
		investments = investments[:len(investments)-1]
	}

	return investments, nil
}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amount := lib.DollarsToCents(jsonCommand.Amount)

	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user.Cents += amount
	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
	}

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quote := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}

func handleBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than buy amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &amountInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks

	buyStackMap.push(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks, moneyToRemove)

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy := buyStackMap.pop(jsonCommand.Userid)
	if nextBuy == nil {
		errorMessage := "No Buy to Commit"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	if user.Cents < nextBuy.cents {
		errorMessage := "Account balance is less than stock cost"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	user.Cents -= nextBuy.cents
	user.Investments = addStock(user.Investments, nextBuy.stockSymbol, nextBuy.numOfStocks)
	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "remove",
		UserID:       user.Command_ID,
		FundsInCents: nextBuy.cents,
	})

	lib.ServerSendOKResponse(conn)
}

func handleCancelBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy := buyStackMap.pop(jsonCommand.Userid)
	if nextBuy == nil {
		errorMessage := "No Buy to Cancel"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than sell amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &amountInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToAdd := quoteInCents * numOfStocks

	sellStackMap.push(jsonCommand.Userid, jsonCommand.StockSymbol, numOfStocks, moneyToAdd)

	lib.ServerSendOKResponse(conn)
}

func handleCommitSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell := sellStackMap.pop(jsonCommand.Userid)
	if nextSell == nil {
		errorMessage := "No Sell to Commit"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	stockAmount := findStockAmount(user.Investments, nextSell.stockSymbol)

	if stockAmount < nextSell.numOfStocks {
		errorMessage := "Not enough stock is owned user to sell"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	investmentsAfterRemove, removeErr := removeStock(user.Investments, nextSell.stockSymbol, nextSell.numOfStocks)
	if removeErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, removeErr.Error())
		return
	}

	user.Investments = investmentsAfterRemove
	user.Cents += nextSell.cents

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, removeErr.Error())
		return
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       user.Command_ID,
		FundsInCents: nextSell.cents,
	})

	lib.ServerSendOKResponse(conn)
}

func handleCancelSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell := sellStackMap.pop(jsonCommand.Userid)
	if nextSell == nil {
		errorMessage := "No Sell to Cancel"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalErrorMessage: errorMessage,
		})
		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	balanceInCents := user.Cents

	if amountInCents > balanceInCents {
		errorMessage := "Account balance is less than trigger amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &balanceInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	_, readErr = dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if readErr == nil {
		errorMessage := "There is an existing trigger, use CANCEL_SET_BUY to update amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalFundsInCents: &balanceInCents,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	} else if readErr != dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user.Cents -= amountInCents
	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	newTrigger := modelsdata.Trigger{jsonCommand.Userid, jsonCommand.StockSymbol, 0, amountInCents, false}
	createErr := dataClient.CreateTrigger(newTrigger)
	if createErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	triggerPriceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, readErr := dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if readErr == dataclient.ErrNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if trigger.Amount_Cents < triggerPriceInCents {
		errorMessage := "Trigger price is too high, no stocks will be able to be bought with current amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	trigger.Price_Cents = triggerPriceInCents
	updateErr := dataClient.UpdateTrigger(trigger)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, readErr := dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if readErr == dataclient.ErrNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr == nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user.Cents += trigger.Amount_Cents

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	deleteErr := dataClient.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if deleteErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
		// TODO handle error here
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, readErr := dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if readErr == nil {
		errorMessage := "There is an existing trigger, use CANCEL_SET_SELL to update amount"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	} else if readErr != dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	newTrigger := modelsdata.Trigger{jsonCommand.Userid, jsonCommand.StockSymbol, 0, amountInCents, true}
	createErr := dataClient.CreateTrigger(newTrigger)
	if createErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	priceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, readErr := dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if readErr == dataclient.ErrNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	if priceInCents > trigger.Amount_Cents {
		errorMessage := "Trigger amount is higher than amount of stocks to sell"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	numOfStocksOwn := findStockAmount(user.Investments, jsonCommand.StockSymbol)
	numOfStocks := trigger.Amount_Cents / priceInCents

	if numOfStocks > numOfStocksOwn {
		errorMessage := "User does not have enough stocks"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	// If there's an existing trigger, give back the reserved stock
	var reservedStocks uint64 = 0
	if trigger.Price_Cents != 0 {
		reservedStocks += trigger.Amount_Cents / trigger.Price_Cents
	}

	if reservedStocks > numOfStocks {
		user.Investments = addStock(user.Investments, jsonCommand.StockSymbol, reservedStocks-numOfStocks)
	} else if reservedStocks < numOfStocks {
		newInvestments, removeErr := removeStock(user.Investments, jsonCommand.StockSymbol, numOfStocks-reservedStocks)
		if removeErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, removeErr.Error())
			return
		}
		user.Investments = newInvestments
	}

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	trigger.Price_Cents = priceInCents
	updateErr = dataClient.UpdateTrigger(trigger)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	trigger, readErr := dataClient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)

	if readErr == dataclient.ErrNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.LogErrorEvent(auditclient.ErrorEventInfo{
			OptionalUserID:       jsonCommand.Userid,
			OptionalStockSymbol:  jsonCommand.StockSymbol,
			OptionalErrorMessage: errorMessage,
		})

		lib.ServerSendResponse(conn, lib.StatusUserError, errorMessage)
		return
	}

	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	deleteErr := dataClient.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if deleteErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
		// TODO handle error here
		return
	}

	if trigger.Price_Cents == 0 {
		// Trigger not set, no need to re-add stock
		lib.ServerSendOKResponse(conn)
		return
	}

	numOfStocks := trigger.Amount_Cents / trigger.Price_Cents
	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user.Investments = addStock(user.Investments, jsonCommand.StockSymbol, numOfStocks)

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleDisplaySummary(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	user, readErr := dataClient.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	triggers, readErr := dataClient.ReadTriggersByUser(user.Command_ID)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	userDisplay := modelsdata.UserDisplayInfo{}
	userDisplay.Cents = user.Cents
	userDisplay.Investments = user.Investments

	// convert triggers to triggerdisplayinfos
	triggerDisplays := []modelsdata.TriggerDisplayInfo{}
	for _, trigger := range triggers {
		triggerDisplays = append(
			triggerDisplays,
			modelsdata.TriggerDisplayInfo{trigger.Stock, trigger.Price_Cents, trigger.Amount_Cents, trigger.Is_Sell},
		)
	}
	userDisplay.Triggers = triggerDisplays

	userDisplayBytes, jsonErr := json.Marshal(userDisplay)
	if jsonErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, jsonErr.Error())
		return
	}
	lib.ServerSendResponse(conn, lib.StatusOk, string(userDisplayBytes))
}
