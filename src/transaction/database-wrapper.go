package main

import (
	"errors"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
	"net"
)

type databaseWrapper struct {
	client net.Conn
}

type stock struct {
	stockSymbol string
	numOfStocks uint64
}

var dataClient = dataclient.DataClient{}

// These two read functions are here to perform some additional error handling
func readUser(userid string) (modelsdata.User, error) {
	user, err := dataClient.ReadUser(userid)
	if err != nil {
		errToReturn := err
		if err == dataclient.ErrNotFound {
			errToReturn = ErrDataNotFound
		}
		return modelsdata.User{}, errToReturn
	}

	return user, nil
}

func readTrigger(userid string, stockSymbol string, isSell bool) (modelsdata.Trigger, error) {
	trigger, err := dataClient.ReadTrigger(userid, stockSymbol, isSell)
	if err != nil {
		errToReturn := err
		if err == dataclient.ErrNotFound {
			errToReturn = ErrDataNotFound
		}
		return modelsdata.Trigger{}, errToReturn
	}

	return trigger, nil
}

func (client *databaseWrapper) userExists(userid string) (bool, error) {
	_, err := dataClient.ReadUser(userid)

	if err != nil {
		if err == dataclient.ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (client *databaseWrapper) createUser(userid string) error {
	user := modelsdata.User{userid, 0, []modelsdata.Investment{}}
	return dataClient.CreateUser(user)
}

func (client *databaseWrapper) addAmount(userid string, cents uint64, auditClient *auditclient.AuditClient) error {
	user, readErr := readUser(userid)
	if readErr != nil {
		return readErr
	}

	user.Cents += cents
	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "add",
		UserID:       userid,
		FundsInCents: cents,
	})

	return nil
}

func (client *databaseWrapper) getBalance(userid string) (uint64, error) {
	user, readErr := readUser(userid)
	if readErr != nil {
		return 0, readErr
	}

	return user.Cents, nil
}

func (client *databaseWrapper) removeAmount(userid string, cents uint64, auditClient *auditclient.AuditClient) error {
	user, readErr := readUser(userid)
	if readErr != nil {
		return readErr
	}

	if user.Cents < cents {
		return errors.New("The user does not have sufficient funds ( " + string(user.Cents) + " ) to remove " + string(cents))
	}

	user.Cents -= cents
	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	auditClient.LogAccountTransaction(auditclient.AccountTransactionInfo{
		Action:       "remove",
		UserID:       userid,
		FundsInCents: cents,
	})

	return nil
}

func (client *databaseWrapper) getStockAmount(userid string, stockSymbol string) (uint64, error) {
	user, readErr := readUser(userid)
	if readErr != nil {
		return 0, readErr
	}

	investmentIndex := -1
	for i, investment := range user.Investments {
		if investment.Stock == stockSymbol {
			investmentIndex = i
		}
	}

	if investmentIndex == -1 {
		return 0, nil
	}

	return user.Investments[investmentIndex].Amount, nil
}

func (client *databaseWrapper) addStock(userid string, stockSymbol string, amount uint64) error {
	user, readErr := readUser(userid)
	if readErr != nil {
		return readErr
	}

	// Find the investment in the user struct and set the amount specified in the params
	investmentIndex := -1
	for i, investment := range user.Investments {
		if investment.Stock == stockSymbol {
			investmentIndex = i
		}
	}

	// If you can't find the investment create a new investment, otherwise add to the existing one
	if investmentIndex == -1 {
		user.Investments = append(user.Investments, modelsdata.Investment{stockSymbol, amount})
	} else {
		user.Investments[investmentIndex].Amount += amount
	}

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (client *databaseWrapper) removeStock(userid string, stockSymbol string, amount uint64) error {
	user, readErr := readUser(userid)
	if readErr != nil {
		return readErr
	}

	investmentIndex := -1
	for i, investment := range user.Investments {
		if investment.Stock == stockSymbol {
			investmentIndex = i
		}
	}

	// Make sure the stock is found
	if investmentIndex == -1 {
		return errors.New("The user with id " + userid + " does not have any of the stock " + stockSymbol)
	}

	// Make sure they have enough stock to remove the amount
	userStockAmount := user.Investments[investmentIndex].Amount
	if userStockAmount < amount {
		return errors.New("The user does not have sufficient stock ( " + string(userStockAmount) + " ) to remove " + string(amount))
	}

	remainingAmount := userStockAmount - amount
	user.Investments[investmentIndex].Amount = remainingAmount

	// If the remaining amount is 0 remove the investment from the user
	if remainingAmount == 0 {
		user.Investments[investmentIndex] = user.Investments[len(user.Investments)-1]
		user.Investments = user.Investments[:len(user.Investments)-1]
	}

	updateErr := dataClient.UpdateUser(user)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (client *databaseWrapper) getTrigger(userid string, stockSymbol string, isSell bool) (modelsdata.Trigger, error) {
	trigger, readErr := readTrigger(userid, stockSymbol, isSell)
	if readErr != nil {
		return modelsdata.Trigger{}, readErr
	}

	return trigger, nil
}

func (client *databaseWrapper) createTrigger(userid string, stockSymbol string, amount_cents uint64, isSell bool) error {
	newTrigger := modelsdata.Trigger{userid, stockSymbol, 0, amount_cents, isSell}
	createErr := dataClient.CreateTrigger(newTrigger)
	if createErr != nil {
		return createErr
	}

	return nil
}

func (client *databaseWrapper) setTriggerPrice(userid string, stockSymbol string, cents uint64, isSell bool) error {
	trigger, readErr := readTrigger(userid, stockSymbol, isSell)
	if readErr != nil {
		return readErr
	}

	trigger.Price_Cents = cents
	updateErr := dataClient.UpdateTrigger(trigger)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (client *databaseWrapper) deleteTrigger(userid string, stockSymbol string, isSell bool) error {
	deleteErr := dataClient.DeleteTrigger(userid, stockSymbol, isSell)
	if deleteErr != nil {
		if deleteErr == dataclient.ErrNotFound {
			return ErrDataNotFound
		}
		return deleteErr
	}

	return nil
}

func (client *databaseWrapper) getTriggers() ([]modelsdata.Trigger, error) {
	triggers, readErr := dataClient.ReadTriggers()
	if readErr != nil {
		return []modelsdata.Trigger{}, readErr
	}

	return triggers, nil
}

func (client *databaseWrapper) getTriggersByUser(userid string) ([]modelsdata.Trigger, error) {
	triggers, readErr := dataClient.ReadTriggersByUser(userid)
	if readErr != nil {
		return []modelsdata.Trigger{}, readErr
	}

	return triggers, nil
}

func (client *databaseWrapper) getStocks(userid string) ([]modelsdata.Investment, error) {
	user, readErr := readUser(userid)
	if readErr != nil {
		return []modelsdata.Investment{}, readErr
	}

	return user.Investments, nil
}
