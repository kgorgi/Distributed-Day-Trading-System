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
type AuditClient struct{}

func unixTimestamp() int32 {
	return int32(time.Now().Unix())
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
	conn, err := net.Dial("tcp", auditServerDockerAddress)
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

// LogUserCommandRequest sends a log of UserCommandType to the audit server
func (client *AuditClient) LogUserCommandRequest(info UserCommandInfo) {
	info.LogType = "UserCommandType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}

// LogQuoteServerResponse sends a UserCommandType to the audit server
func (client *AuditClient) LogQuoteServerResponse(info QuoteServerResponseInfo) {
	info.LogType = "QuoteServerType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}

// LogAccountTransaction sends a log of UserCommandType to the audit server
func (client *AuditClient) LogAccountTransaction(info AccountTransactionInfo) {
	info.LogType = "AccountTransactionType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}

// LogSystemEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogSystemEvent(info SystemEventInfo) {
	info.LogType = "SystemEventType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}

// LogErrorEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogErrorEvent(info ErrorEventInfo) {
	info.LogType = "ErrorEventType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}

// LogDebugEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogDebugEvent(info DebugEventInfo) {
	info.LogType = "DebutEventType"
	info.Timestamp = unixTimestamp()

	client.sendLogs(info)
}
