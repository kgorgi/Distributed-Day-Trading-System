package main

import (
	"sync/atomic"
	"time"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// TODO This is not thread safe :(
type stack struct {
	items  []*reserve
	userID string
	isBuy  bool
}

type reserve struct {
	stockSymbol string
	numOfStocks uint64
	cents       uint64
	valid       uint64
	timer       *time.Timer

	// Used to return the go routine
	done chan struct{}
}

var buyStack = make(map[string]*stack)
var sellStack = make(map[string]*stack)

func getBuyStack(userID string) *stack {
	if buyStack[userID] == nil {
		stack := new(stack)
		stack.isBuy = true
		stack.items = make([]*reserve, 0)
		stack.userID = userID
		buyStack[userID] = stack
	}

	return buyStack[userID]
}

func getSellStack(userID string) *stack {
	if sellStack[userID] == nil {
		stack := new(stack)
		stack.isBuy = false
		stack.items = make([]*reserve, 0)
		stack.userID = userID
		sellStack[userID] = stack
	}

	return sellStack[userID]
}

func createReseve(stockSymbol string, numOfStocks uint64, cents uint64) *reserve {
	var instance *reserve
	instance = new(reserve)
	instance.stockSymbol = stockSymbol
	instance.numOfStocks = numOfStocks
	instance.cents = cents
	return instance
}

func (stack *stack) push(newItem *reserve, auditClient *auditclient.AuditClient) {
	newItem.valid = 0
	newItem.timer = time.NewTimer(time.Second * 60)
	newItem.done = make(chan struct{})

	go func() {
		select {
		case <-newItem.timer.C:
			// Timer reached buy/sell cancelled
			atomic.AddUint64(&newItem.valid, 1)

			if stack.isBuy {
				err := dataConn.addAmount(stack.userID, newItem.cents, auditClient)
				if err != nil {
					// TODO
				}
			} else {
				err := dataConn.addStock(stack.userID, newItem.stockSymbol, newItem.numOfStocks)
				if err != nil {
					// TODO
				}
			}
		case <-newItem.done:
			// Timer cancelled early
			return
		}
	}()

	stack.items = append(stack.items, newItem)
}

func (stack *stack) pop() *reserve {
	numOfItems := len(stack.items)
	if numOfItems == 0 {
		return nil
	}

	n := numOfItems - 1
	var nextItem *reserve = stack.items[n]
	nextItem.timer.Stop()
	close(nextItem.done)

	// Find first valid item or end of list
	for atomic.LoadUint64(&nextItem.valid) != 0 && n > 0 {
		stack.items[n] = nil
		stack.items = stack.items[:n]
		n = n - 1
		nextItem = stack.items[n]
		nextItem.timer.Stop()
		close(nextItem.done)
	}

	stack.items[n] = nil
	stack.items = stack.items[:n]
	if atomic.LoadUint64(&nextItem.valid) == 0 {
		return nextItem
	}

	// n = 0
	return nil
}
