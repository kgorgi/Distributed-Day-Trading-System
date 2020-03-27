package main

import (
	"encoding/json"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func processCommand(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	valid := validateUser(conn, jsonCommand, &auditClient)
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
		handleDisplaySummary(conn, jsonCommand, &auditClient)
	default:
		lib.ServerSendResponse(conn, lib.StatusUserError, "Invalid command")
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

func validateUser(conn net.Conn, commandJSON CommandJSON, auditClient *auditclient.AuditClient) bool {
	// Validate user exists
	_, err := dataclient.ReadUser(commandJSON.Userid, auditClient)
	if err != nil && err != dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return false
	}

	// If the user exist return true
	if err == nil {
		return true
	}

	// If the user doesn't exist, and the command is not ADD return false
	if commandJSON.Command != "ADD" {
		lib.ServerSendResponse(conn, lib.StatusUserError, "User does not exist")
		return false
	}

	// If the user doens't exist and the command is ADD create a new user
	newUser := modelsdata.User{
		Command_ID:  commandJSON.Userid,
		Cents:       0,
		Investments: []modelsdata.Investment{},
		Buys:        []modelsdata.Reserve{},
		Sells:       []modelsdata.Reserve{},
	}
	createErr := dataclient.CreateUser(newUser, auditClient)
	if createErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
		return false
	}

	return true
}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amount := lib.DollarsToCents(jsonCommand.Amount)
	addErr := dataclient.UpdateUser(jsonCommand.Userid, "", 0, int(amount), auditClient)
	if addErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, addErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quote := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, false, auditClient)
	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}

func handleBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, true, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than buy amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks

	pushErr := dataclient.PushUserBuy(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToRemove, numOfStocks, auditClient)
	if pushErr == dataclient.ErrNotFound {
		errorMessage := "The specified user does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
	}

	if pushErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, pushErr.Error())
	}

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy, popErr := dataclient.PopUserBuy(jsonCommand.Userid, auditClient)
	if popErr == dataclient.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any valid buys"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	buyErr := dataclient.UpdateUser(jsonCommand.Userid, nextBuy.Stock, int(nextBuy.Num_Stocks), int(nextBuy.Cents)*-1, auditClient)
	if buyErr != nil {
		auditClient.LogErrorEvent(buyErr.Error())

		if buyErr == dataclient.ErrNotFound {
			errorMessage := "The specified user either does not exist or does not have sufficient funds to remove " + strconv.FormatUint(nextBuy.Cents, 10) + " cents"
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
			return
		}

		lib.ServerSendResponse(conn, lib.StatusSystemError, buyErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, popErr := dataclient.PopUserBuy(jsonCommand.Userid, auditClient)
	if popErr == dataclient.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any buys to cancel"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	quoteInCents := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, true, auditClient)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than sell amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToAdd := quoteInCents * numOfStocks

	pushErr := dataclient.PushUserSell(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToAdd, numOfStocks, auditClient)
	if pushErr == dataclient.ErrNotFound {
		errorMessage := "The specified user does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
	}

	if pushErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, pushErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCommitSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell, popErr := dataclient.PopUserSell(jsonCommand.Userid, auditClient)
	if popErr == dataclient.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any valid sells"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	sellErr := dataclient.UpdateUser(jsonCommand.Userid, nextSell.Stock, int(nextSell.Num_Stocks)*-1, int(nextSell.Cents), auditClient)
	if sellErr != nil {
		auditClient.LogErrorEvent(sellErr.Error())

		if sellErr == dataclient.ErrNotFound {
			errorMessage := "Either the specified user does not exist, or they do not have a sufficient amount of stock"
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
			return
		}

		lib.ServerSendResponse(conn, lib.StatusSystemError, sellErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, popErr := dataclient.PopUserSell(jsonCommand.Userid, auditClient)
	if popErr == dataclient.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any sells to cancel"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)
	user, readErr := dataclient.ReadUser(jsonCommand.Userid, auditClient)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	balanceInCents := user.Cents

	if amountInCents > balanceInCents {
		errorMessage := "Account balance is less than trigger amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	existingTrigger, getTriggerErr := dataclient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false, auditClient)
	if getTriggerErr != nil && getTriggerErr != dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, getTriggerErr.Error())
		return
	}

	var existingAmount uint64 = 0
	if getTriggerErr == dataclient.ErrNotFound {
		newTrigger := modelsdata.Trigger{
			User_Command_ID:    jsonCommand.Userid,
			Stock:              jsonCommand.StockSymbol,
			Price_Cents:        0,
			Amount_Cents:       amountInCents,
			Is_Sell:            false,
			Transaction_Number: auditClient.TransactionNum,
		}

		createErr := dataclient.CreateTrigger(newTrigger, auditClient)
		if createErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
			return
		}

	} else {
		if amountInCents < existingTrigger.Price_Cents {
			errorMessage := "An existing trigger on this stock has a higher trigger price than the set amount"
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
			return
		}

		existingAmount = existingTrigger.Amount_Cents
		updateErr := dataclient.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, false, amountInCents, auditClient)
		if updateErr == dataclient.ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The trigger was fired before the update could occur")
			return
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			return
		}
	}

	updateErr := dataclient.UpdateUser(jsonCommand.Userid, "", 0, int(existingAmount)-int(amountInCents), auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	triggerPriceInCents := lib.DollarsToCents(jsonCommand.Amount)
	updateErr := dataclient.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, false, triggerPriceInCents, auditClient)
	if updateErr == dataclient.ErrNotFound {
		errorMessage := "Either the trigger doesn't exist, or the specified price is too high, no stocks will be able to be bought with current amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	deletedTrigger, deleteErr := dataclient.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false, auditClient)

	// If the trigger doesn't exist or has been deleted by the time this is executing
	if deleteErr == dataclient.ErrNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if deleteErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
		return
	}

	// If the trigger was successfully deleted, then give the triggers resererved money back to the user
	updateErr := dataclient.UpdateUser(jsonCommand.Userid, "", 0, int(deletedTrigger.Amount_Cents), auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {

	existingTrigger, getTriggerErr := dataclient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, auditClient)
	if getTriggerErr != nil && getTriggerErr != dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, getTriggerErr.Error())
		return
	}
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	if getTriggerErr == dataclient.ErrNotFound {
		newTrigger := modelsdata.Trigger{
			User_Command_ID:    jsonCommand.Userid,
			Stock:              jsonCommand.StockSymbol,
			Price_Cents:        0,
			Amount_Cents:       amountInCents,
			Is_Sell:            true,
			Transaction_Number: auditClient.TransactionNum,
		}
		err := dataclient.CreateTrigger(newTrigger, auditClient)
		if err != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
			return
		}

		lib.ServerSendOKResponse(conn)
		return
	}

	if amountInCents < existingTrigger.Price_Cents {
		errorMessage := "An existing trigger on this stock has a higher trigger price than the set amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	//Update the trigger and handle the case where the trigger is fired off
	err := dataclient.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, true, amountInCents, auditClient)
	if err == dataclient.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusNotFound, "The trigger was fired before the update could happen")
		return
	}

	if err != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, err.Error())
		return
	}

	if existingTrigger.Price_Cents <= 0 {
		lib.ServerSendOKResponse(conn)
		return
	}

	//Now that we know the trigger was successfully updated we can update the user
	reservedStock := existingTrigger.Amount_Cents / existingTrigger.Price_Cents
	newStock := amountInCents / existingTrigger.Price_Cents

	updateUserErr := dataclient.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStock)-int(newStock), 0, auditClient)
	if updateUserErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateUserErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	priceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, readErr := dataclient.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, auditClient)
	if readErr == dataclient.ErrNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user, readErr := dataclient.ReadUser(jsonCommand.Userid, auditClient)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}
	numOfStocksOwn := findStockAmount(user.Investments, jsonCommand.StockSymbol)
	numOfStocks := trigger.Amount_Cents / priceInCents

	if numOfStocks > numOfStocksOwn {
		errorMessage := "User does not have enough stocks"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	updateErr := dataclient.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, true, priceInCents, auditClient)
	if updateErr == dataclient.ErrNotFound {
		errorMessage := "The specified trigger has fired, or the trigger amount is higher than amount of stocks to sell"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	var reservedStocks uint64 = 0
	if trigger.Price_Cents != 0 {
		reservedStocks += trigger.Amount_Cents / trigger.Price_Cents
	}

	userUpdateErr := dataclient.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStocks)-int(numOfStocks), 0, auditClient)
	if userUpdateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, userUpdateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	deletedTrigger, deleteErr := dataclient.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, auditClient)

	// If the trigger doesn't exist, or was fired before this was by the time this part of the function executes
	if deleteErr == dataclient.ErrNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if deleteErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
		return
	}

	if deletedTrigger.Price_Cents == 0 {
		// Trigger not set, no need to re-add stock
		lib.ServerSendOKResponse(conn)
		return
	}

	// If the trigger was successfully deleted, then we add back the corresponding stock
	numOfStocks := deletedTrigger.Amount_Cents / deletedTrigger.Price_Cents
	updateErr := dataclient.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(numOfStocks), 0, auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleDisplaySummary(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	user, readErr := dataclient.ReadUser(jsonCommand.Userid, auditClient)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	triggers, readErr := dataclient.ReadTriggersByUser(user.Command_ID, auditClient)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	userDisplay := modelsdata.UserDisplayInfo{}
	userDisplay.Cents = user.Cents
	userDisplay.Investments = user.Investments

	triggerDisplays := []modelsdata.TriggerDisplayInfo{}
	for _, trigger := range triggers {
		triggerDisplays = append(
			triggerDisplays,
			modelsdata.TriggerDisplayInfo{
				Stock:        trigger.Stock,
				Price_Cents:  trigger.Price_Cents,
				Amount_Cents: trigger.Amount_Cents,
				Is_Sell:      trigger.Is_Sell,
			},
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
