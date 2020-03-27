package main

import (
	"encoding/json"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
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

func findStockAmount(investments []data.Investment, stockSymbol string) uint64 {
	for _, investment := range investments {
		if investment.Stock == stockSymbol {
			return investment.Amount
		}
	}
	return 0
}

func validateUser(conn net.Conn, commandJSON CommandJSON, auditClient *auditclient.AuditClient) bool {
	// Validate user exists
	_, err := data.ReadUser(commandJSON.Userid)
	if err != nil && err != data.ErrNotFound {
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
	newUser := data.User{
		Command_ID:  commandJSON.Userid,
		Cents:       0,
		Investments: []data.Investment{},
		Buys:        []data.Reserve{},
		Sells:       []data.Reserve{},
	}
	createErr := data.CreateUser(newUser)
	if createErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, createErr.Error())
		return false
	}

	return true
}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amount := lib.DollarsToCents(jsonCommand.Amount)
	addErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(amount), auditClient)
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

	pushErr := data.PushUserReserve(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToRemove, numOfStocks, false)
	if pushErr == data.ErrNotFound {
		errorMessage := "The specified user does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
	}

	if pushErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, pushErr.Error())
	}

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy, popErr := data.PopUserReserve(jsonCommand.Userid, false)
	if popErr == data.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any valid buys"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	buyErr := data.UpdateUser(jsonCommand.Userid, nextBuy.Stock, int(nextBuy.Num_Stocks), int(nextBuy.Cents)*-1, auditClient)
	if buyErr != nil {
		auditClient.LogErrorEvent(buyErr.Error())

		if buyErr == data.ErrNotFound {
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
	_, popErr := data.PopUserReserve(jsonCommand.Userid, false)
	if popErr == data.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any buys to cancel"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
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

	pushErr := data.PushUserReserve(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToAdd, numOfStocks, true)
	if pushErr == data.ErrNotFound {
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
	nextSell, popErr := data.PopUserReserve(jsonCommand.Userid, true)
	if popErr == data.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any valid sells"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, popErr.Error())
		return
	}

	sellErr := data.UpdateUser(jsonCommand.Userid, nextSell.Stock, int(nextSell.Num_Stocks)*-1, int(nextSell.Cents), auditClient)
	if sellErr != nil {
		auditClient.LogErrorEvent(sellErr.Error())

		if sellErr == data.ErrNotFound {
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
	_, popErr := data.PopUserReserve(jsonCommand.Userid, true)
	if popErr == data.ErrNotFound {
		errorMessage := "The specified user either does not exist or does not any sells to cancel"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
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
	user, readErr := data.ReadUser(jsonCommand.Userid)
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

	existingTrigger, getTriggerErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)
	if getTriggerErr != nil && getTriggerErr != data.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, getTriggerErr.Error())
		return
	}

	var existingAmount uint64 = 0
	if getTriggerErr == data.ErrNotFound {
		newTrigger := data.Trigger{
			User_Command_ID:    jsonCommand.Userid,
			Stock:              jsonCommand.StockSymbol,
			Price_Cents:        0,
			Amount_Cents:       amountInCents,
			Is_Sell:            false,
			Transaction_Number: auditClient.TransactionNum,
		}

		createErr := data.CreateTrigger(newTrigger)
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
		updateErr := data.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, false, amountInCents)
		if updateErr == data.ErrNotFound {
			lib.ServerSendResponse(conn, lib.StatusNotFound, "The trigger was fired before the update could occur")
			return
		}

		if updateErr != nil {
			lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
			return
		}
	}

	updateErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(existingAmount)-int(amountInCents), auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	triggerPriceInCents := lib.DollarsToCents(jsonCommand.Amount)
	updateErr := data.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, false, triggerPriceInCents)
	if updateErr == data.ErrNotFound {
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
	deletedTrigger, deleteErr := data.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false)

	// If the trigger doesn't exist or has been deleted by the time this is executing
	if deleteErr == data.ErrNotFound {
		errorMessage := "Trigger does not exist"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if deleteErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, deleteErr.Error())
		return
	}

	// If the trigger was successfully deleted, then give the triggers resererved money back to the user
	updateErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(deletedTrigger.Amount_Cents), auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {

	existingTrigger, getTriggerErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if getTriggerErr != nil && getTriggerErr != data.ErrNotFound {
		lib.ServerSendResponse(conn, lib.StatusSystemError, getTriggerErr.Error())
		return
	}
	amountInCents := lib.DollarsToCents(jsonCommand.Amount)

	if getTriggerErr == data.ErrNotFound {
		newTrigger := data.Trigger{
			User_Command_ID:    jsonCommand.Userid,
			Stock:              jsonCommand.StockSymbol,
			Price_Cents:        0,
			Amount_Cents:       amountInCents,
			Is_Sell:            true,
			Transaction_Number: auditClient.TransactionNum,
		}
		err := data.CreateTrigger(newTrigger)
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
	err := data.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, true, amountInCents)
	if err == data.ErrNotFound {
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

	updateUserErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStock)-int(newStock), 0, auditClient)
	if updateUserErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateUserErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	priceInCents := lib.DollarsToCents(jsonCommand.Amount)
	trigger, readErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)
	if readErr == data.ErrNotFound {
		errorMessage := "Trigger amount has not been set"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	user, readErr := data.ReadUser(jsonCommand.Userid)
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

	updateErr := data.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, true, priceInCents)
	if updateErr == data.ErrNotFound {
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

	userUpdateErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStocks)-int(numOfStocks), 0, auditClient)
	if userUpdateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, userUpdateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	deletedTrigger, deleteErr := data.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true)

	// If the trigger doesn't exist, or was fired before this was by the time this part of the function executes
	if deleteErr == data.ErrNotFound {
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
	updateErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(numOfStocks), 0, auditClient)
	if updateErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, updateErr.Error())
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleDisplaySummary(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	user, readErr := data.ReadUser(jsonCommand.Userid)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	triggers, readErr := data.ReadTriggersByUser(user.Command_ID)
	if readErr != nil {
		lib.ServerSendResponse(conn, lib.StatusSystemError, readErr.Error())
		return
	}

	userDisplay := data.UserDisplayInfo{}
	userDisplay.Cents = user.Cents
	userDisplay.Investments = user.Investments

	triggerDisplays := []data.TriggerDisplayInfo{}
	for _, trigger := range triggers {
		triggerDisplays = append(
			triggerDisplays,
			data.TriggerDisplayInfo{
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
