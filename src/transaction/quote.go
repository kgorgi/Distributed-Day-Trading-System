package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/resolveurl"
)

type quote struct {
	amount      uint64
	stockSymbol string
	timestamp   uint64
	cryptokey   string
	mutex       sync.RWMutex
}

type quoteCache struct {
	quotes map[string]*quote
	mutex  sync.RWMutex
}

var cache = quoteCache{
	quotes: make(map[string]*quote),
}

const quoteServerAddress = "192.168.1.100:4443"

var sixtySecondsInMs = uint64(60 * time.Second / time.Millisecond)

// GetQuote returns a quote from the quote server
func GetQuote(
	stockSymbol string,
	userID string,
	noCache bool,
	auditClient *auditclient.AuditClient) uint64 {

	q := cache.getQuote(stockSymbol)
	return q.getCents(userID, noCache, auditClient)
}

func (qc *quoteCache) createQuote(stockSymbol string) {
	q := new(quote)
	q.stockSymbol = stockSymbol

	cache.mutex.Lock()
	if cache.quotes[stockSymbol] != nil {
		cache.mutex.Unlock()
		return
	}

	cache.quotes[stockSymbol] = q
	cache.mutex.Unlock()
}

func (qc *quoteCache) getQuote(stockSymbol string) *quote {
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

func (q *quote) valid() bool {
	return (lib.GetUnixTimestamp() - q.timestamp) < sixtySecondsInMs
}

func (q *quote) getCents(
	userID string,
	noCache bool,
	auditClient *auditclient.AuditClient) uint64 {

	q.mutex.RLock()
	if q.valid() && !noCache {
		// Use Cache
		amount := q.amount
		timestamp := q.timestamp
		cryptokey := q.cryptokey
		q.mutex.RUnlock()

		message := "Retrieved quote from cache (timestamp: " + strconv.FormatUint(timestamp, 10) +
			", cryptokey: " + cryptokey + ")"

		auditClient.LogDebugEvent(auditclient.DebugEventInfo{
			OptionalUserID:       userID,
			OptionalFundsInCents: &amount,
			OptionalDebugMessage: message,
		})

		return amount
	}
	q.mutex.RUnlock()

	// Update Cache
	amount := q.updateQuote(userID, auditClient)
	return amount
}

func (q *quote) updateQuote(
	userID string,
	auditClient *auditclient.AuditClient) uint64 {

	var address = resolveurl.MockQuoteServerAddress()
	if lib.UseLabQuoteServer() {
		address = quoteServerAddress
	}

	// Establish Connection to Quote Server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Could not connect to quote server")
		return 0
	}

	// Send Request
	payload := q.stockSymbol + "," + userID + "\n"
	_, err = conn.Write([]byte(payload))
	if err != nil {
		log.Fatalln("Failed to send request to quote server")
	}

	// Receive Response
	rawResponse, err := bufio.NewReader(conn).ReadString('\n')
	rawResponse = strings.TrimRight(rawResponse, "\n")

	if err != nil {
		log.Fatalln("Failed to recieve response to quote server")
		return 0
	}

	// Process Data
	data := strings.Split(rawResponse, ",")

	if len(data) < 4 {
		log.Fatalln("Quote server response is incorrect")
		return 0
	}

	timestamp, err := strconv.ParseUint(data[3], 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse timestamp from quote server")
		return 0
	}

	cents := lib.DollarsToCents(data[0])
	auditClient.LogQuoteServerResponse(auditclient.QuoteServerResponseInfo{
		QuoteServerTime: timestamp,
		UserID:          userID,
		PriceInCents:    cents,
		StockSymbol:     q.stockSymbol,
		CryptoKey:       data[4],
	})

	q.mutex.Lock()
	q.amount = cents
	q.timestamp = lib.GetUnixTimestamp()
	q.cryptokey = data[4]
	q.mutex.Unlock()

	return cents
}
