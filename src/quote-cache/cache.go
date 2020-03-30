package main

import (
	"strconv"
	"sync"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/quote"
	"extremeWorkload.com/daytrader/lib/serverurls"
)

type quoteValue struct {
	amount      uint64
	stockSymbol string
	timestamp   uint64
	cryptokey   string
	mutex       sync.RWMutex
}

type quoteCache struct {
	quotes map[string]*quoteValue
	mutex  sync.RWMutex
}

var cache = quoteCache{
	quotes: make(map[string]*quoteValue),
}

var quoteServerAddress = serverurls.Env.LegacyQuoteServer

var sixtySecondsInMs = uint64(60 * time.Second / time.Millisecond)

// GetQuote returns a quote from the quote server
func GetQuote(
	stockSymbol string,
	userID string,
	noCache bool,
	auditClient *auditclient.AuditClient) (uint64, error) {

	q := cache.getQuote(stockSymbol)
	return q.getCents(userID, noCache, auditClient)
}

func (qc *quoteCache) createQuote(stockSymbol string) {
	q := new(quoteValue)
	q.stockSymbol = stockSymbol

	cache.mutex.Lock()
	if cache.quotes[stockSymbol] != nil {
		cache.mutex.Unlock()
		return
	}

	cache.quotes[stockSymbol] = q
	cache.mutex.Unlock()
}

func (qc *quoteCache) getQuote(stockSymbol string) *quoteValue {
	qc.mutex.RLock()
	q := cache.quotes[stockSymbol]
	cache.mutex.RUnlock()

	if q == nil {
		qc.createQuote(stockSymbol)

		qc.mutex.RLock()
		q = cache.quotes[stockSymbol]
		cache.mutex.RUnlock()
	}

	return q
}

func (q *quoteValue) valid() bool {
	return (lib.GetUnixTimestamp() - q.timestamp) < sixtySecondsInMs
}

func (q *quoteValue) getCents(
	userID string,
	noCache bool,
	auditClient *auditclient.AuditClient) (uint64, error) {

	q.mutex.RLock()
	if q.valid() && !noCache {
		// Use Cache
		amount := q.amount
		timestamp := q.timestamp
		cryptokey := q.cryptokey
		q.mutex.RUnlock()

		message := "Retrieved quote from cache (timestamp: " + strconv.FormatUint(timestamp, 10) +
			", cryptokey: " + cryptokey + ")"

		auditClient.LogDebugEvent(message)

		return amount, nil
	}
	q.mutex.RUnlock()

	// Update Cache
	amount, err := q.updateQuote(userID, auditClient)
	return amount, err
}

func (q *quoteValue) updateQuote(
	userID string,
	auditClient *auditclient.AuditClient) (uint64, error) {

	cents, cryptokey, err := quote.Request(q.stockSymbol, userID, auditClient)
	if err != nil {
		return 0, err
	}

	q.mutex.Lock()
	q.amount = cents
	q.timestamp = lib.GetUnixTimestamp()
	q.cryptokey = cryptokey
	q.mutex.Unlock()

	return cents, nil
}
