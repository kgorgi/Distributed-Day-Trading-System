package main

import (
	"bufio"
	"log"
	"net"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

// GetQuote returns a quote from the quote server
func GetQuote(
	stockSymbol string,
	userID string,
	auditClient auditclient.AuditClient) uint64 {

	// Establish Connection to Quote Server
	conn, err := net.Dial("tcp", ":5000")
	if err != nil {
		log.Fatalln("Could not connect to quote server")
		return 0
	}

	// Send Request
	payload := stockSymbol + "," + userID
	_, err = conn.Write([]byte(payload))
	if err != nil {
		log.Fatalln("Failed to send request to quote server")
		return 0
	}

	// Receive Response
	rawResponse, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatalln("Failed to send recieve response to quote server")
		return 0
	}

	// Process Data
	data := strings.Split(rawResponse, ",")

	if len(data) < 4 {
		log.Fatalln("Quote server response is incorrect")
		return 0
	}

	// timestamp, err := strconv.ParseUint(data[3], 10, 4)
	// if err != nil {
	// 	log.Fatalln("Failed to parse timestamp from quote server")
	// 	return 0
	// }

	cents := lib.DollarsToCents(data[0])

	// auditClient.LogQuoteServerResponse(auditclient.QuoteServerResponseInfo{
	// 	QuoteServerTime: timestamp,
	// 	UserID:          userID,
	// 	PriceInCents:    cents,
	// 	StockSymbol:     stockSymbol,
	// 	CryptoKey:       data[4],
	// })

	return cents
}
