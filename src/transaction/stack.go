package main

import (
	"sync/atomic"
	"time"
)

type stack struct {
	items  []*reserve
	userID string
}

type reserve struct {
	stockSymbol string
	numOfStocks uint64
	cents       uint64
	valid       uint64
	timer       *time.Timer
}

var buyStack = make(map[string]*stack)
var sellStack = make(map[string]*stack)

func getBuyStack(userID string) *stack {
	if buyStack[userID] == nil {
		stack := new(stack)
		stack.items = make([]*reserve, 0)
		stack.userID = userID
		buyStack[userID] = stack
	}

	return buyStack[userID]
}

func getSellStack(userID string) *stack {
	if sellStack[userID] == nil {
		stack := new(stack)
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

func (stack *stack) push(newItem *reserve) {
	stack.items = append(stack.items, newItem)
	newItem.valid = 0

	newItem.timer = time.NewTimer(time.Second * 60)

	go func() {
		<-newItem.timer.C
		atomic.AddUint64(&newItem.valid, 1)
		err := dataConn.addAmount(stack.userID, newItem.cents)
		if err != nil {
			// TODO
		}
	}()
}

func (stack *stack) pop() *reserve {
	numOfItems := len(stack.items)
	if numOfItems == 0 {
		return nil
	}

	n := numOfItems - 1
	var nextItem *reserve = stack.items[n]
	nextItem.timer.Stop()
	for atomic.LoadUint64(&nextItem.valid) == 0 && n > 0 {
		stack.items[n] = nil
		n = n - 1
		nextItem = stack.items[n]
		nextItem.timer.Stop()
	}

	if atomic.LoadUint64(&nextItem.valid) == 0 {
		return nextItem
	}

	// n = 0
	stack.items[n] = nil
	return nil
}
