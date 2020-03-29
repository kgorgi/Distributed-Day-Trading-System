package main

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/serverurls"
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

func (q *quote) updateQuote(
	userID string,
	auditClient *auditclient.AuditClient) (uint64, error) {

	// Establish Connection to Quote Server
	conn, err := net.Dial("tcp", quoteServerAddress)
	if err != nil {
		return 0, errors.New("Failed to contact quote server " + err.Error())
	}

	// Send Request
	payload := q.stockSymbol + "," + userID + "\n"
	_, err = conn.Write([]byte(payload))
	if err != nil {
		conn.Close()
		return 0, errors.New("Failed to send request to quote server " + err.Error())
	}

	// Receive Response
	rawResponse, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Close()
		return 0, errors.New("Failed to recieve response to quote server " + err.Error())
	}

	conn.Close()

	// Process Data
	rawResponse = strings.TrimRight(rawResponse, "\n")
	data := strings.Split(rawResponse, ",")

	if len(data) < 4 {
		return 0, errors.New("Quote server response is incorrect")
	}

	if lib.IsLab {
		if data[1] != q.stockSymbol {
			return 0, errors.New("Response's stock symbol is incorrect")
		}

		if data[2] != userID {
			return 0, errors.New("Response's userid is incorrect")
		}
	}

	timestamp, err := strconv.ParseUint(data[3], 10, 64)
	if err != nil {
		return 0, errors.New("Failed to parse timestamp from quote server " + err.Error())
	}

	cents := lib.DollarsToCents(data[0])
	auditClient.LogQuoteServerResponse(cents, q.stockSymbol, userID, timestamp, data[4])

	q.mutex.Lock()
	q.amount = cents
	q.timestamp = lib.GetUnixTimestamp()
	q.cryptokey = data[4]
	q.mutex.Unlock()

	return cents, nil
}
