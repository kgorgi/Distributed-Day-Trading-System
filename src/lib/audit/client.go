package auditclient

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"extremeWorkload.com/daytrader/lib"
)

const auditServerDockerAddress = "audit-server:5002"
const auditServerLocalAddress = "localhost:5002"

// AuditClient send requests to the audit server
type AuditClient struct {
	Server string
}

// LogUserCommandRequest sends a log of UserCommandType to the audit server
func (client *AuditClient) LogUserCommandRequest(info UserCommandInfo) {
	var internalInfo = client.generateInternalInfo("UserCommandType")
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
	var internalInfo = client.generateInternalInfo("QuoteServerType")
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
	var internalInfo = client.generateInternalInfo("AccountTransactionType")
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
	var internalInfo = client.generateInternalInfo("SystemEventType")
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
	var internalInfo = client.generateInternalInfo("ErrorEventType")
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
	var internalInfo = client.generateInternalInfo("DebugEventType")
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

	// Establish Connection to Audit Server
	conn, err := net.Dial("tcp", auditServerLocalAddress)
	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return
	}

	// Send Payload
	status, message, err := lib.ClientSendRequest(conn, payload)

	conn.Close()

	if err != nil {
		log.Println("Connection Error: " + err.Error())
		return
	}

	if status != lib.StatusOk {
		log.Println("Response Error: Status " + string(status) + message)
		return
	}
}

func (client *AuditClient) generateInternalInfo(logType string) InternalLogInfo {
	return InternalLogInfo{
		LogType:   logType,
		Timestamp: int32(time.Now().Unix()),
		Server:    client.Server,
	}
}
