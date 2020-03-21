package auditclient

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"strconv"

	"extremeWorkload.com/daytrader/lib/serverurls"

	"extremeWorkload.com/daytrader/lib"
)

// AuditClient send requests to the audit server
type AuditClient struct {
	Server         string
	TransactionNum uint64
	Command        string
}

// DumpLogAll get all logs from audit server
func (client *AuditClient) DumpLogAll() (int, string, error) {
	return client.DumpLog("")
}

// DumpLog get all logs from audit server
func (client *AuditClient) DumpLog(userID string) (int, string, error) {
	return client.sendRequest("DUMPLOG|" + userID)
}

// LogUserCommandRequest sends a log of UserCommandType to the audit server
// and returns the global transaction number given by the audit server.
// If the client cannot contact the audit server then a puesdorandom transaction
// number will be generated
func (client *AuditClient) LogUserCommandRequest(info UserCommandInfo) uint64 {
	var internalInfo = client.generateInternalInfo("userCommand", true)
	payload := struct {
		*InternalLogInfo
		*UserCommandInfo
	}{
		&internalInfo,
		&info,
	}

	// Convert JSON to Payload
	jsonText, err := json.Marshal(payload)
	if err != nil {
		log.Println("JSON stringify error: " + err.Error())
		return client.setRandomTransactionNum()
	}

	payloadStr := "USERLOG" + "|" + string(jsonText)
	status, message, err := client.sendRequest(payloadStr)
	handleRequestFailure(status, message, err, payloadStr)

	if status != lib.StatusOk || err != nil {
		return client.setRandomTransactionNum()
	}

	result, err := strconv.ParseUint(message, 10, 64)
	if err != nil {
		log.Println("Audit Client: Failed to Parse Result: " + message)
		return client.setRandomTransactionNum()
	}

	client.TransactionNum = result

	return result
}

// LogQuoteServerResponse sends a QuoteServerType to the audit server
func (client *AuditClient) LogQuoteServerResponse(
	priceInCents uint64,
	stockSymbol string,
	userID string,
	quoteServerTime uint64,
	cryptoKey string,
) {
	var internalInfo = client.generateInternalInfo("quoteServer", false)
	var transactionInfo = QuoteServerResponseInfo{
		PriceInCents:    priceInCents,
		StockSymbol:     stockSymbol,
		UserID:          userID,
		QuoteServerTime: quoteServerTime,
		CryptoKey:       cryptoKey,
	}

	payload := struct {
		*InternalLogInfo
		*QuoteServerResponseInfo
	}{
		&internalInfo,
		&transactionInfo,
	}

	client.sendLogs(payload)
}

// LogAccountTransaction sends a log of AccountTransactionType to the audit server
func (client *AuditClient) LogAccountTransaction(
	userID string,
	fundsInCents int64,
) {
	var action = "add"
	amount := fundsInCents
	if fundsInCents < 0 {
		action = "remove"
		amount = -fundsInCents
	}

	var internalInfo = client.generateInternalInfo("accountTransaction", false)
	var transactionInfo = AccountTransactionInfo{
		Action:       action,
		UserID:       userID,
		FundsInCents: uint64(amount),
	}

	payload := struct {
		*InternalLogInfo
		*AccountTransactionInfo
	}{
		&internalInfo,
		&transactionInfo,
	}

	client.sendLogs(payload)
}

// LogSystemEvent sends a log of SystemEventType to the audit server
func (client *AuditClient) LogSystemEvent() {
	var internalInfo = client.generateInternalInfo("systemEvent", true)
	client.sendLogs(internalInfo)
}

// LogErrorEvent sends a log of ErrorEventType to the audit server
func (client *AuditClient) LogErrorEvent(errorMessage string) {
	var internalInfo = client.generateInternalInfo("errorEvent", true)
	var errorInfo = ErrorEventInfo{
		ErrorMessage: errorMessage,
	}

	payload := struct {
		*InternalLogInfo
		*ErrorEventInfo
	}{
		&internalInfo,
		&errorInfo,
	}

	client.sendLogs(payload)
}

// SendServerResponseWithErrorEvent sends a log of ErrorEventType to the audit server
// and sends the error response to the client
func (client *AuditClient) SendServerResponseWithErrorEvent(conn net.Conn, status int, errorMessage string) {
	client.LogErrorEvent(errorMessage)
	lib.ServerSendResponse(conn, status, errorMessage)
}

// LogDebugEvent sends a log of DebugEventType to the audit server
func (client *AuditClient) LogDebugEvent(debugMessage string) {
	var internalInfo = client.generateInternalInfo("debugEvent", true)
	var debugInfo = DebugEventInfo{
		DebugMessage: debugMessage,
	}

	payload := struct {
		*InternalLogInfo
		*DebugEventInfo
	}{
		&internalInfo,
		&debugInfo,
	}

	client.sendLogs(payload)
}

func (client *AuditClient) sendLogs(data interface{}) {
	// Convert JSON to Payload
	jsonText, err := json.Marshal(data)
	if err != nil {
		log.Println("JSON stringify error: " + err.Error())
		return
	}

	payload := "LOG" + "|" + string(jsonText)

	status, message, err := client.sendRequest(payload)
	handleRequestFailure(status, message, err, payload)
}

func (client *AuditClient) generateInternalInfo(logType string, withCommand bool) InternalLogInfo {
	var internalInfo = InternalLogInfo{
		LogType:        logType,
		Timestamp:      lib.GetUnixTimestamp(),
		Server:         client.Server,
		TransactionNum: client.TransactionNum,
	}

	if withCommand && client.Command != "" {
		internalInfo.Command = client.Command
	}

	return internalInfo
}

func (client *AuditClient) sendRequest(payload string) (int, string, error) {
	// Send Payload
	status, message, err := lib.ClientSendRequest(serverurls.Env.AuditServer, payload)
	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return -1, "", err
	}
	return status, message, nil
}

func (client *AuditClient) setRandomTransactionNum() uint64 {
	num := rand.Uint64()
	client.TransactionNum = num
	return num
}

func handleRequestFailure(status int, message string, err error, payload string) {
	if err != nil {
		log.Println("Audit Client Connection Error: " + err.Error() + " Payload: " + payload)
	} else if status != lib.StatusOk {
		log.Println("Audit Client Response Error: Status " + strconv.Itoa(status) + " " +
			message +
			" Payload: " + payload)
	}
}
