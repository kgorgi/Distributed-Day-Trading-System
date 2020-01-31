package main

import (
    "net"
    "errors"
    dataclient "extremeWorkload.com/daytrader/lib/data"
    modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

type databaseWrapper struct {
    client net.Conn
}

type stock struct {
    stockSymbol string
    numOfStocks uint64
}

var dataClient = dataclient.DataClient{}

// IsUserExist check if user is in db
func (client *databaseWrapper) userExists(userid string) (bool, error) {
    
    //Read User should return an error if there is no matching user
    _, err := dataClient.ReadUser(userid);
    if err != nil {
        return false, err
    }
    
    return true, nil
}

func (client *databaseWrapper) createUser(userid string) error {
    user := modelsdata.User{userid, 0, []modelsdata.Investment{}}
    return dataClient.CreateUser(user)
}

func (client *databaseWrapper) addAmount(userid string, cents uint64) error {
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return readErr
    }

    user.Cents += cents
    updateErr := dataClient.UpdateUser(user)
    if updateErr != nil {
        return updateErr
    }
    
    return nil
}

func (client *databaseWrapper) getBalance(userid string) (uint64, error) {
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return 0, readErr
    }

    return user.Cents, nil
}

func (client *databaseWrapper) removeAmount(userid string, cents uint64) error {
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return readErr
    }

    if user.Cents < cents {
        return errors.New("The user does not have sufficient funds ( " + string(user.Cents) + " ) to remove " + string(cents));
    }

    updateErr := dataClient.UpdateUser(user)
    if updateErr != nil {
        return updateErr
    }
    
    return nil
}

func (client *databaseWrapper) getStockAmount(userid string, stockSymbol string) (uint64, error) {
    //return stocks[stockSymbol], nil
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return 0, readErr
    }

    var amount uint64
    for _, investment := range user.Investments {
        if(investment.Stock == stockSymbol) {
            amount = investment.Amount
        }
    }
    return amount, nil
}

func (client *databaseWrapper) addStock(userid string, stockSymbol string, amount uint64) error {
    //read the client first
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return readErr
    }

    //find the investment in the user struct and set the amount specified in the params
    var investmentIndex int
    for i, investment := range user.Investments {
        if(investment.Stock == stockSymbol) {
            investmentIndex = i
        }
    }
    user.Investments[investmentIndex].Amount += amount

    //update the user in the db
    updateErr := dataClient.UpdateUser(user)
    if updateErr != nil {
        return updateErr
    }

    return nil
}

func (client *databaseWrapper) removeStock(userid string, stockSymbol string, amount uint64) error {
    //read the client first
    user, readErr := dataClient.ReadUser(userid);
    if readErr != nil {
        return readErr
    }

    //find the investment in the user struct and set the amount specified in the params
    var investmentIndex int
    for i, investment := range user.Investments {
        if(investment.Stock == stockSymbol) {
            investmentIndex = i
        }
    }

    //check to see if the user has enough stock to remove
    userStockAmount := user.Investments[investmentIndex].Amount
    if userStockAmount < amount {
        return errors.New("The user does not have sufficient stock ( " + string(userStockAmount) + " ) to remove " + string(amount));
    }

    //update the user in the db
    user.Investments[investmentIndex].Amount -= amount
    updateErr := dataClient.UpdateUser(user)
    if updateErr != nil {
        return updateErr
    }

    return nil
}

func (client *databaseWrapper) getTrigger(userid string, stockSymbol string, isSell bool) (modelsdata.Trigger, error) {
	trigger, readErr := dataClient.ReadTrigger(userid, stockSymbol, isSell);
	if readErr != nil {
		return modelsdata.Trigger{}, readErr
	}

	return trigger, nil
}

func (client *databaseWrapper) createTrigger(userid string, stockSymbol string, amount_cents uint64, isSell bool) error {
	newTrigger := modelsdata.Trigger{ userid, stockSymbol, 0, amount_cents, isSell }
	createErr := dataClient.CreateTrigger(newTrigger);
	if createErr != nil {
		return createErr
	}

	return nil
}

//probably should be called set trigger price
func (client *databaseWrapper) setTriggerAmount(userid string, stockSymbol string, cents uint64, isSell bool) error {
	trigger, readErr := dataClient.ReadTrigger(userid, stockSymbol, isSell)
	if readErr != nil {
		return readErr
	}

	trigger.Price_Cents = cents;
	updateErr := dataClient.UpdateTrigger(trigger)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (client *databaseWrapper) deleteTrigger(userid string, stockSymbol string, isSell bool) error {
	deleteErr := dataClient.DeleteTrigger(userid, stockSymbol, isSell)
	if deleteErr != nil {
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

func (client *databaseWrapper) getStocks(userid string) ([]modelsdata.Investment, error) {
	user, readErr := dataClient.ReadUser(userid)
	if readErr != nil {
		return []modelsdata.Investment{}, readErr
	}

	return user.Investments, nil
}


