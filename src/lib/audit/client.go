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
	conn net.Conn
}

func (client *AuditClient) connect() error {
	var err error
	client.conn, err = net.Dial("tcp", auditServerDockerAddress)
	return err
}

func (client *AuditClient) disconnect() error {
	return client.conn.Close()
}

func unixTimestamp() int32 {
	return int32(time.Now().Unix())
}

func (client *AuditClient) sendLogs(data interface{}) {
	jsonText, err := json.Marshal(data)
	if err != nil {
		log.Fatal("JSON stringify error: " + err.Error())
		return
	}

	payload := "LOG|" + string(jsonText)

	client.connect()

	status, message, err := lib.ClientSendRequest(client.conn, payload)

	client.disconnect()

	if err != nil {
		log.Println("Network Error: " + err.Error())
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
