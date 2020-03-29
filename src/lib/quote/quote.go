package quote

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
	"extremeWorkload.com/daytrader/lib/serverurls"
)

var quoteServerAddress = serverurls.Env.LegacyQuoteServer

// Request makes a request and processes responses from the quote server
// Returns the amount of cents and the cryptokey
func Request(
	stockSymbol string,
	userID string,
	auditClient *auditclient.AuditClient) (uint64, string, error) {
	// Establish Connection to Quote Server
	conn, err := net.Dial("tcp", quoteServerAddress)
	if err != nil {
		return 0, "", errors.New("Failed to contact quote server " + err.Error())
	}

	// Send Request
	payload := stockSymbol + "," + userID + "\n"
	_, err = conn.Write([]byte(payload))
	if err != nil {
		conn.Close()
		return 0, "", errors.New("Failed to send request to quote server " + err.Error())
	}

	// Receive Response
	rawResponse, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Close()
		return 0, "", errors.New("Failed to recieve response to quote server " + err.Error())
	}

	conn.Close()

	// Process Data
	rawResponse = strings.TrimRight(rawResponse, "\n")
	data := strings.Split(rawResponse, ",")

	if len(data) < 4 {
		return 0, "", errors.New("Quote server response is incorrect")
	}

	if lib.IsLab {
		if data[1] != stockSymbol {
			return 0, "", errors.New("Response's stock symbol is incorrect")
		}

		if data[2] != userID {
			return 0, "", errors.New("Response's userid is incorrect")
		}
	}

	timestamp, err := strconv.ParseUint(data[3], 10, 64)
	if err != nil {
		return 0, "", errors.New("Failed to parse timestamp from quote server " + err.Error())
	}

	cents := lib.DollarsToCents(data[0])
	cyptokey := data[4]
	auditClient.LogQuoteServerResponse(cents, stockSymbol, userID, timestamp, cyptokey)

	return cents, cyptokey, nil
}
