package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type StackMap struct {
	stacks map[string]*stack
	mutex  sync.RWMutex
}

// TODO This is not thread safe :(
type stack struct {
	items  []*reserve
	userID string
	mutex  sync.Mutex
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

var buyStackMap = StackMap{
	stacks: make(map[string]*stack),
}

var sellStackMap = StackMap{
	stacks: make(map[string]*stack),
}

func createReserve(stockSymbol string, numOfStocks uint64, cents uint64) *reserve {
	var newItem *reserve
	newItem = new(reserve)
	newItem.stockSymbol = stockSymbol
	newItem.numOfStocks = numOfStocks
	newItem.cents = cents
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

	return newItem

}

func (stackMap *StackMap) createStack(userid string) {
	stack := new(stack)
	stack.items = make([]*reserve, 0)
	stack.userID = userid

	stackMap.mutex.Lock()
	if stackMap.stacks[userid] != nil {
		stackMap.mutex.Unlock()
		return
	}

	stackMap.stacks[userid] = stack
	stackMap.mutex.Unlock()

}

func (stackMap *StackMap) getStack(userid string) *stack {
	stackMap.mutex.RLock()
	stack := stackMap.stacks[userid]
	stackMap.mutex.RUnlock()

	if stack == nil {
		stackMap.createStack(userid)

		stackMap.mutex.RLock()
		stack = stackMap.stacks[userid]
		stackMap.mutex.RUnlock()
	}

	return stack
}

func (stackMap *StackMap) push(userid string, stockSymbol string, numOfStocks uint64, cents uint64) {
	stack := stackMap.getStack(userid)
	stack.mutex.Lock()
	newItem := createReserve(stockSymbol, numOfStocks, cents)
	stack.items = append(stack.items, newItem)
	stack.mutex.Unlock()
}

func (stackMap *StackMap) pop(userid string) *reserve {
	stack := stackMap.getStack(userid)

	stack.mutex.Lock()
	numOfItems := len(stack.items)
	if numOfItems == 0 {
		stack.mutex.Unlock()
		return nil
	}

	n := numOfItems - 1
	var topOfStack *reserve = stack.items[n]

	// Is first item valid
	if !(atomic.LoadUint64(&topOfStack.notValid) == 1) {
		topOfStack.timer.Stop()
		close(topOfStack.done)
		stack.items[n] = nil
		stack.items = stack.items[:n]

		stack.mutex.Unlock()
		return topOfStack
	}

	// Dis-regard stack all invalid
	stack.items = make([]*reserve, 0)
	stack.mutex.Unlock()
	return nil
}
