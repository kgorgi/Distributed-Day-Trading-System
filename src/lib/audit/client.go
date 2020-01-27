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
	payload := client.attachAdditionalInfo(info, "UserCommandType")
	client.sendLogs(payload)
}

// LogQuoteServerResponse sends a UserCommandType to the audit server
func (client *AuditClient) LogQuoteServerResponse(info QuoteServerResponseInfo) {
	payload := client.attachAdditionalInfo(info, "QuoteServerType")
	client.sendLogs(payload)
}

// LogAccountTransaction sends a log of UserCommandType to the audit server
func (client *AuditClient) LogAccountTransaction(info AccountTransactionInfo) {
	payload := client.attachAdditionalInfo(info, "AccountTransactionType")
	client.sendLogs(payload)
}

// LogSystemEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogSystemEvent(info SystemEventInfo) {
	payload := client.attachAdditionalInfo(info, "SystemEventType")
	client.sendLogs(payload)
}

// LogErrorEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogErrorEvent(info ErrorEventInfo) {
	payload := client.attachAdditionalInfo(info, "ErrorEventType")
	client.sendLogs(payload)
}

// LogDebugEvent sends a log of UserCommandType to the audit server
func (client *AuditClient) LogDebugEvent(info DebugEventInfo) {
	payload := client.attachAdditionalInfo(info, "DebutEventType")
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

func (client *AuditClient) attachAdditionalInfo(data interface{}, logType string) interface{} {
	tmp := struct {
		data      interface{}
		LogType   string `json:"logType" bson:"logType"`
		Timestamp int32  `json:"timestamp" bson:"timestamp"`
		Server    string `json:"server" bson:"server"`
	}{
		data:      data,
		LogType:   logType,
		Timestamp: int32(time.Now().Unix()),
		Server:    client.Server,
	}

	return tmp
}
