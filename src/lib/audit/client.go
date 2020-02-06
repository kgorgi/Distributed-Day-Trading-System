package auditclient

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/resolveurl"
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
func (client *AuditClient) LogUserCommandRequest(info UserCommandInfo) {
	var internalInfo = client.generateInternalInfo("userCommand", true)
	payload := struct {
		*InternalLogInfo
		*UserCommandInfo
	}{
		&internalInfo,
		&info,
	}

	client.sendLogs(payload)
}

// LogQuoteServerResponse sends a UserCommandType to the audit server
func (client *AuditClient) LogQuoteServerResponse(info QuoteServerResponseInfo) {
	var internalInfo = client.generateInternalInfo("quoteServer", false)
	payload := struct {
		*InternalLogInfo
		*QuoteServerResponseInfo
	}{
		&internalInfo,
		&info,
	}
	client.sendLogs(payload)
}

// LogAccountTransaction sends a log of UserCommandType to the audit server
func (client *AuditClient) LogAccountTransaction(info AccountTransactionInfo) {
	var internalInfo = client.generateInternalInfo("accountTransaction", false)
	payload := struct {
		*InternalLogInfo
		*AccountTransactionInfo
	}{
		&internalInfo,
		&info,
	}
	client.sendLogs(payload)
}

// LogSystemEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogSystemEvent(info SystemEventInfo) {
	var internalInfo = client.generateInternalInfo("systemEvent", true)
	payload := struct {
		*InternalLogInfo
		*SystemEventInfo
	}{
		&internalInfo,
		&info,
	}
	client.sendLogs(payload)
}

// LogErrorEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogErrorEvent(info ErrorEventInfo) {
	var internalInfo = client.generateInternalInfo("errorEvent", true)
	payload := struct {
		*InternalLogInfo
		*ErrorEventInfo
	}{
		&internalInfo,
		&info,
	}
	client.sendLogs(payload)
}

// LogDebugEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogDebugEvent(info DebugEventInfo) {
	var internalInfo = client.generateInternalInfo("debugEvent", true)
	payload := struct {
		*InternalLogInfo
		*DebugEventInfo
	}{
		&internalInfo,
		&info,
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

	payload := "LOG|" + string(jsonText)

	client.sendRequest(payload)
}

func (client *AuditClient) generateInternalInfo(logType string, withCommand bool) InternalLogInfo {
	var internalInfo = InternalLogInfo{
		LogType:        logType,
		Timestamp:      uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
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
	conn, err := net.Dial("tcp", resolveurl.AuditServerAddress())
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
