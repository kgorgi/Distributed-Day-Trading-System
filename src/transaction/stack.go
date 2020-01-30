package main

import (
	"sync/atomic"
	"time"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

type stack struct {
	items  []*reserve
	userID string
	isBuy  bool
}

type reserve struct {
	stockSymbol string
	numOfStocks uint64
	cents       uint64
	notValid    uint64
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

func createReserve(stockSymbol string, numOfStocks uint64, cents uint64) *reserve {
	var instance *reserve
	instance = new(reserve)
	instance.stockSymbol = stockSymbol
	instance.numOfStocks = numOfStocks
	instance.cents = cents
	return instance
}

func (stack *stack) push(newItem *reserve, auditClient *auditclient.AuditClient) {
	newItem.notValid = 0
	newItem.timer = time.NewTimer(time.Second * 60)
	newItem.done = make(chan struct{})

	go func() {
		select {
		case <-newItem.timer.C:
			// Timer reached buy/sell cancelled
			atomic.AddUint64(&newItem.notValid, 1)
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
	var topOfStack *reserve = stack.items[n]

	// Is first item valid
	if atomic.LoadUint64(&topOfStack.notValid) != 0 {
		topOfStack.timer.Stop()
		close(topOfStack.done)
		stack.items[n] = nil
		stack.items = stack.items[:n]
		return topOfStack
	}

	// Dis-regard stack all invalid
	stack.items = make([]*reserve, 0)
	return nil
}
