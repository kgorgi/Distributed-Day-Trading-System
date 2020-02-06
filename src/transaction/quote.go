package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/resolveurl"
)

const quoteServerAddress = "192.168.1.100:4443"

// GetQuote returns a quote from the quote server
func GetQuote(
	stockSymbol string,
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
	payload := stockSymbol + "," + userID + "\n"
	_, err = conn.Write([]byte(payload))
	if err != nil {
		log.Fatalln("Failed to send request to quote server")
		return 0
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
		StockSymbol:     stockSymbol,
		CryptoKey:       data[4],
	})

	return cents
}
