package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/transaction/data"
)

func processCommand(conn net.Conn, jsonCommand CommandJSON, auditClient auditclient.AuditClient) {
	// Validate user exists
	_, err := data.ReadUser(jsonCommand.Userid, context.TODO())
	if err != nil && err != data.ErrNotFound {
		errorMessage := "Database failure: " + err.Error()
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	// Create a new user if ADD command and it does not exist
	if err == data.ErrNotFound && jsonCommand.Command == "ADD" {
		newUser := data.User{
			Command_ID:  jsonCommand.Userid,
			Cents:       0,
			Investments: []data.Investment{},
			Buys:        []data.Reserve{},
			Sells:       []data.Reserve{},
		}

		createErr := data.CreateUser(newUser)
		if createErr != nil {
			errorMessage := "Failed to create user error"
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
			return
		}
	}

	// If the user doesn't exist, and the command is not ADD
	if err == data.ErrNotFound && jsonCommand.Command != "ADD" {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, "User does not exist")
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
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, "Invalid command")
	}

}

func handleAdd(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	amount, _ := lib.DollarsToCents(jsonCommand.Amount)
	addErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(amount), context.TODO(), auditClient)
	if addErr != nil {
		errorMessage := "Failed to add user " + addErr.Error()
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleQuote(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quote, err := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, false, auditClient)
	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, err.Error())
		return
	}

	dollars := lib.CentsToDollars(quote)
	lib.ServerSendResponse(conn, lib.StatusOk, dollars)
}

func handleBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quoteInCents, err := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, true, auditClient)
	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, err.Error())
		return
	}

	amountInCents, _ := lib.DollarsToCents(jsonCommand.Amount)
	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than buy amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToRemove := quoteInCents * numOfStocks

	pushErr := data.PushUserReserve(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToRemove, numOfStocks, false)
	if pushErr != nil {
		errorMessage := "Failed to push buy request on stack " + pushErr.Error()
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCommitBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextBuy, popErr := data.PopUserReserve(jsonCommand.Userid, false)
	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		errorMessage := failedToPopStackMessage("buy", popErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	buyErr := data.UpdateUser(jsonCommand.Userid, nextBuy.Stock, int(nextBuy.Num_Stocks), int(nextBuy.Cents)*-1, context.TODO(), auditClient)

	if buyErr == data.ErrNotFound {
		errorMessage := "The user does not have sufficient funds to remove " + strconv.FormatUint(nextBuy.Cents, 10) + " cents"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if buyErr != nil {
		errorMessage := failedToUpdateUserMessage(buyErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, popErr := data.PopUserReserve(jsonCommand.Userid, false)
	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		errorMessage := failedToPopStackMessage("buy", popErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	quoteInCents, err := GetQuote(jsonCommand.StockSymbol, jsonCommand.Userid, true, auditClient)
	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, err.Error())
		return
	}
	amountInCents, _ := lib.DollarsToCents(jsonCommand.Amount)

	if quoteInCents > amountInCents {
		errorMessage := "Quote price is higher than sell amount"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	numOfStocks := amountInCents / quoteInCents
	moneyToAdd := quoteInCents * numOfStocks

	pushErr := data.PushUserReserve(jsonCommand.Userid, jsonCommand.StockSymbol, moneyToAdd, numOfStocks, true)
	if pushErr != nil {
		errorMessage := "Failed to push sell request onto stack: " + err.Error()
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCommitSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	nextSell, popErr := data.PopUserReserve(jsonCommand.Userid, true)
	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		errorMessage := failedToPopStackMessage("sell", popErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	sellErr := data.UpdateUser(jsonCommand.Userid, nextSell.Stock, int(nextSell.Num_Stocks)*-1, int(nextSell.Cents), context.TODO(), auditClient)
	if sellErr == data.ErrNotFound {
		errorMessage := "The user does not have a sufficient amount of stock"
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if sellErr != nil {
		errorMessage := failedToUpdateUserMessage(sellErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	_, popErr := data.PopUserReserve(jsonCommand.Userid, true)
	if popErr == data.ErrEmptyStack {
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, popErr.Error())
		return
	}

	if popErr != nil {
		errorMessage := failedToPopStackMessage("sell", popErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleSetBuyAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		user, readErr := data.ReadUser(jsonCommand.Userid, ctx)
		if readErr != nil {
			errorMessage := failedToReadUserMessage(readErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		amountInCents, _ := lib.DollarsToCents(jsonCommand.Amount)
		balanceInCents := user.Cents
		if amountInCents > balanceInCents {
			errorMessage := "Account balance is less than trigger amount"
			return errors.New(errorMessage), lib.StatusSystemError
		}

		existingTrigger, getTriggerErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false, ctx)
		if getTriggerErr != nil && getTriggerErr != data.ErrNotFound {
			errorMessage := failedToReadTriggerMessage(getTriggerErr)
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
			return errors.New(errorMessage), lib.StatusSystemError
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

			createErr := data.CreateTrigger(newTrigger, ctx)
			if createErr != nil {
				errorMessage := failedToCreateTriggerMessage(createErr)
				return errors.New(errorMessage), lib.StatusSystemError

			}

			updateErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(existingAmount)-int(amountInCents), ctx, auditClient)
			if updateErr != nil {
				errorMessage := failedToUpdateUserMessage(updateErr)
				return errors.New(errorMessage), lib.StatusSystemError

			}
		} else {
			if amountInCents < existingTrigger.Price_Cents {
				errorMessage := "An existing trigger on this stock has a higher trigger price than the set amount"
				auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
				return errors.New(errorMessage), lib.StatusUserError
			}

			existingAmount = existingTrigger.Amount_Cents

			updateErr := data.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, false, amountInCents, ctx)
			if updateErr == data.ErrNotFound {
				errorMessage := "The trigger was fired before the update could occur"
				return errors.New(errorMessage), lib.StatusNotFound
			}

			if updateErr != nil {
				errorMessage := failedToUpdateTriggerAmount(updateErr)
				return errors.New(errorMessage), lib.StatusSystemError

			}

			updateErr = data.UpdateUser(jsonCommand.Userid, "", 0, int(existingAmount)-int(amountInCents), ctx, auditClient)
			if updateErr != nil {
				errorMessage := failedToUpdateUserMessage(updateErr)
				return errors.New(errorMessage), lib.StatusSystemError
			}
		}

		return nil, lib.StatusOk
	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendOKResponse(conn)
	}
}

func handleSetBuyTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	triggerPriceInCents, _ := lib.DollarsToCents(jsonCommand.Amount)
	updateErr := data.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, false, triggerPriceInCents, context.TODO())
	if updateErr == data.ErrNotFound {
		errorMessage := "Either the trigger doesn't exist or the specified price is too high. No stocks will be able to be bought with current amount."
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
		return
	}

	if updateErr != nil {
		errorMessage := failedToUpdateTriggerPrice(updateErr)
		auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
		return
	}

	lib.ServerSendOKResponse(conn)
}

func handleCancelSetBuy(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		deletedTrigger, deleteErr := data.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, false, ctx)

		// If the trigger doesn't exist or has been deleted by the time this is executing
		if deleteErr == data.ErrNotFound {
			errorMessage := "Trigger does not exist"
			return errors.New(errorMessage), lib.StatusUserError
		}

		if deleteErr != nil {
			errorMessage := failedToCancelTrigger(deleteErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		// If the trigger was successfully deleted, then give the triggers resererved money back to the user
		updateErr := data.UpdateUser(jsonCommand.Userid, "", 0, int(deletedTrigger.Amount_Cents), ctx, auditClient)
		if updateErr != nil {
			errorMessage := failedToUpdateUserMessage(updateErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		return nil, lib.StatusOk
	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendOKResponse(conn)
	}
}

func handleSetSellAmount(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		existingTrigger, getTriggerErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, ctx)
		if getTriggerErr != nil && getTriggerErr != data.ErrNotFound {
			errorMessage := failedToReadTriggerMessage(getTriggerErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		amountInCents, _ := lib.DollarsToCents(jsonCommand.Amount)
		if getTriggerErr == data.ErrNotFound {
			newTrigger := data.Trigger{
				User_Command_ID:    jsonCommand.Userid,
				Stock:              jsonCommand.StockSymbol,
				Price_Cents:        0,
				Amount_Cents:       amountInCents,
				Is_Sell:            true,
				Transaction_Number: auditClient.TransactionNum,
			}

			err := data.CreateTrigger(newTrigger, context.TODO())
			if err != nil {
				errorMessage := failedToCreateTriggerMessage(err)
				return errors.New(errorMessage), lib.StatusSystemError
			}
		} else {
			if amountInCents < existingTrigger.Price_Cents {
				errorMessage := "An existing trigger on this stock has a higher trigger price than the set amount"
				return errors.New(errorMessage), lib.StatusUserError
			}

			//Update the trigger and handle the case where the trigger is fired off
			err := data.UpdateTriggerAmount(jsonCommand.Userid, jsonCommand.StockSymbol, true, amountInCents, ctx)
			if err == data.ErrNotFound {
				errorMessage := "The trigger was fired before the update could happen"
				return errors.New(errorMessage), lib.StatusNotFound
			}

			if err != nil {
				errorMessage := failedToUpdateTriggerAmount(err)
				return errors.New(errorMessage), lib.StatusSystemError
			}

			if existingTrigger.Price_Cents > 0 {
				//Now that we know the trigger was successfully updated we can update the user
				reservedStock := existingTrigger.Amount_Cents / existingTrigger.Price_Cents
				newStock := amountInCents / existingTrigger.Price_Cents

				updateUserErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStock)-int(newStock), 0, ctx, auditClient)
				if updateUserErr != nil {
					errorMessage := failedToUpdateUserMessage(updateUserErr)
					return errors.New(errorMessage), lib.StatusSystemError
				}
			}
		}
		return nil, lib.StatusOk

	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendOKResponse(conn)
	}
}

func handleSetSellTrigger(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		trigger, readErr := data.ReadTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, ctx)
		if readErr == data.ErrNotFound {
			errorMessage := "Trigger amount has not been set"
			return errors.New(errorMessage), lib.StatusUserError
		}

		if readErr != nil {
			errorMessage := failedToReadTriggerMessage(readErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		user, readErr := data.ReadUser(jsonCommand.Userid, ctx)
		if readErr != nil {
			errorMessage := failedToReadUserMessage(readErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		numOfStocksOwn := findStockAmount(user.Investments, jsonCommand.StockSymbol)
		priceInCents, _ := lib.DollarsToCents(jsonCommand.Amount)
		numOfStocks := trigger.Amount_Cents / priceInCents

		if numOfStocks > numOfStocksOwn {
			errorMessage := "User does not have enough stocks"
			return errors.New(errorMessage), lib.StatusUserError
		}

		updateErr := data.UpdateTriggerPrice(jsonCommand.Userid, jsonCommand.StockSymbol, true, priceInCents, ctx)
		if updateErr == data.ErrNotFound {
			errorMessage := "The specified trigger has fired or the trigger amount is higher than amount of stocks to sell"
			return errors.New(errorMessage), lib.StatusUserError
		}

		if updateErr != nil {
			errorMessage := failedToUpdateTriggerPrice(updateErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		var reservedStocks uint64 = 0
		if trigger.Price_Cents != 0 {
			reservedStocks += trigger.Amount_Cents / trigger.Price_Cents
		}

		userUpdateErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(reservedStocks)-int(numOfStocks), 0, ctx, auditClient)
		if userUpdateErr != nil {
			errorMessage := failedToUpdateUserMessage(userUpdateErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}

		return nil, lib.StatusOk
	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendOKResponse(conn)
	}
}

func handleCancelSetSell(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		deletedTrigger, deleteErr := data.DeleteTrigger(jsonCommand.Userid, jsonCommand.StockSymbol, true, ctx)

		// If the trigger doesn't exist, or was fired before this was by the time this part of the function executes
		if deleteErr == data.ErrNotFound {
			errorMessage := "Trigger does not exist"
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusUserError, errorMessage)
			return errors.New(errorMessage), lib.StatusUserError
		}

		if deleteErr != nil {
			errorMessage := failedToCancelTrigger(deleteErr)
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
			return errors.New(errorMessage), lib.StatusUserError
		}

		if deletedTrigger.Price_Cents == 0 {
			// Trigger not set, no need to re-add stock
			return nil, lib.StatusOk
		}

		// If the trigger was successfully deleted, then we add back the corresponding stock
		numOfStocks := deletedTrigger.Amount_Cents / deletedTrigger.Price_Cents
		updateErr := data.UpdateUser(jsonCommand.Userid, jsonCommand.StockSymbol, int(numOfStocks), 0, ctx, auditClient)
		if updateErr != nil {
			errorMessage := failedToUpdateUserMessage(updateErr)
			return errors.New(errorMessage), lib.StatusSystemError
		}
		return nil, lib.StatusOk
	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendOKResponse(conn)
	}
}

func handleDisplaySummary(conn net.Conn, jsonCommand CommandJSON, auditClient *auditclient.AuditClient) {
	var resultBytes []byte
	err, status := data.ExecuteTransaction(func(ctx context.Context) (error, int) {
		user, readErr := data.ReadUser(jsonCommand.Userid, ctx)
		if readErr != nil {
			errorMessage := failedToReadUserMessage(readErr)
			return errors.New(errorMessage), lib.StatusSystemError

		}

		triggers, readErr := data.ReadTriggersByUser(user.Command_ID)
		if readErr != nil {
			errorMessage := failedToReadTriggerMessage(readErr)
			auditClient.SendServerResponseWithErrorEvent(conn, lib.StatusSystemError, errorMessage)
			return errors.New(errorMessage), lib.StatusSystemError
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

		var jsonErr error
		resultBytes, jsonErr = json.Marshal(userDisplay)
		if jsonErr != nil {
			errorMessage := "Failed to Marshal JSON " + jsonErr.Error()
			return errors.New(errorMessage), lib.StatusSystemError
		}

		return nil, lib.StatusOk
	})

	if err != nil {
		auditClient.SendServerResponseWithErrorEvent(conn, status, err.Error())
	} else {
		lib.ServerSendResponse(conn, lib.StatusOk, string(resultBytes))
	}
}
