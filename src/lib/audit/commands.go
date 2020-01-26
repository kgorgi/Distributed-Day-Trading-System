package commands

import (
	"encoding/json"
	"time"
)

// Log audit message baseclass
type logBase struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
}

// UserCommandLog audit message for user commands
type userCommandLog struct {
	*logBase
	Command  string `json:"command" bson:"command"`
	Username string `json:"username,omitempty" bson:"username"`
	Filename string `json:"filename,omitempty" bson:"filename,omitempty"`
	Funds    int    `json:"funds,omitempty" bson:"funds,omitempty"`
}

// CreateLogSystemEvent for logType property
func CreateLogSystemEvent(
	server string,
	transactionNum int,
	command string) string {
	datetime := int32(time.Now().Unix())

	base := logBase{
		LogType:        "UserCommandType",
		Timestamp:      datetime,
		Server:         server,
		TransactionNum: transactionNum,
	}

	data := userCommandLog{
		logBase: &base,
		Command: command,
	}

	jsonText, _ := json.Marshal(data)

	payload := "LOG|" + string(jsonText)

	return payload
}
