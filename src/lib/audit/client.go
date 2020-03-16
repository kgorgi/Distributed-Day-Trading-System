package auditclient

import (
	"encoding/json"
	"log"
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
func (client *AuditClient) DumpLogAll() (string, error) {
	return client.DumpLog("")
}

// DumpLog get all logs from audit server
func (client *AuditClient) DumpLog(userID string) (string, error) {
	_, message, err := client.sendRequest("DUMPLOG|" + userID)
	return message, err
}

// LogUserCommandRequest sends a log of UserCommandType to the audit server
func (client *AuditClient) LogUserCommandRequest(info UserCommandInfo) uint64 {
	var internalInfo = client.generateInternalInfo("userCommand", true)
	payload := struct {
		*InternalLogInfo
		*UserCommandInfo
	}{
		&internalInfo,
		&info,
	}

	status, message, err := client.sendLogs(payload, true)
	if err != nil {
		log.Fatalln(err)
	}

	if status != lib.StatusOk {
		log.Fatalln("Status was not okay: " + strconv.FormatInt(int64(status), 10))
	}

	result, err := strconv.ParseUint(message, 10, 64)
	if err != nil {
		log.Fatalln(err)
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
	client.sendLogs(payload, false)
}

// LogAccountTransaction sends a log of AccountTransactionType to the audit server
func (client *AuditClient) LogAccountTransaction(
	action string,
	userID string,
	fundsInCents uint64,
) {
	var internalInfo = client.generateInternalInfo("accountTransaction", false)
	var transactionInfo = AccountTransactionInfo{
		Action:       action,
		UserID:       userID,
		FundsInCents: fundsInCents,
	}

	payload := struct {
		*InternalLogInfo
		*AccountTransactionInfo
	}{
		&internalInfo,
		&transactionInfo,
	}
	client.sendLogs(payload, false)
}

// LogSystemEvent sends a log of SystemEventType to the audit server
func (client *AuditClient) LogSystemEvent() {
	var internalInfo = client.generateInternalInfo("systemEvent", true)
	client.sendLogs(internalInfo, false)
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
	client.sendLogs(payload, false)
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
	client.sendLogs(payload, false)
}

func (client *AuditClient) sendLogs(data interface{}, isUser bool) (int, string, error) {
	// Convert JSON to Payload
	jsonText, err := json.Marshal(data)
	if err != nil {
		log.Println("JSON stringify error: " + err.Error())
		return -1, "", nil
	}

	var requestType = "LOG"
	if isUser {
		requestType = "USERCOMMAND"
	}
	payload := requestType + "|" + string(jsonText)

	return client.sendRequest(payload)
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
	// Establish Connection to Audit Server
	conn, err := net.Dial("tcp", serverurls.Env.AuditServer)
	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return -1, "", err
	}

	// Send Payload
	status, message, err := lib.ClientSendRequest(conn, payload)

	conn.Close()

	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return -1, "", err
	}

	if status != lib.StatusOk {
		log.Println("Response Error: Status " + strconv.Itoa(status) + " " + message)
		return status, message, nil
	}

	return status, message, nil
}
