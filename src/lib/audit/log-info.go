package auditclient

// UserCommandInfo audit message for user commands
type UserCommandInfo struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Command        string `json:"command" bson:"command"`
	Username       string `json:"username,omitempty" bson:"username"`
	Filename       string `json:"filename,omitempty" bson:"filename,omitempty"`
	Funds          *int   `json:"funds,omitempty" bson:"funds,omitempty"`
}

// QuoteServerResponseInfo audit message for quote server responses
type QuoteServerResponseInfo struct {
	LogType         string `json:"logType" bson:"logType"`
	Timestamp       int32  `json:"timestamp" bson:"timestamp"`
	Server          string `json:"server" bson:"server"`
	TransactionNum  int    `json:"transactionNum" bson:"transactionNum"`
	Price           int    `json:"price" bson:"price"`
	StockSymbol     string `json:"stockSymbol" bson:"stockSymbol"`
	Username        string `json:"username" bson:"username"`
	QuoteServerTime uint   `json:"quoteServerTime" bson:"quoteServerTime"`
	CryptoKey       string `json:"cryptoKey" bson:"cryptoKey"`
}

// AccountTransactionInfo audit message for account transactions
type AccountTransactionInfo struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Action         string `json:"action" bson:"action"`
	Username       string `json:"username" bson:"username"`
	Funds          int    `json:"funds" bson:"funds"`
}

// SystemEventInfo audit message for any system events
type SystemEventInfo struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Command        string `json:"command" bson:"command"`
	Username       string `json:"username,omitempty" bson:"username,omitempty"`
	StockSymbol    string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	Filename       string `json:"filename,omitempty" bson:"filename,omitempty"`
	Funds          *int   `json:"funds,omitempty" bson:"funds,omitempty"`
}

// ErrorEventInfo audit message for any system error events
type ErrorEventInfo struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Command        string `json:"command" bson:"command"`
	Username       string `json:"username,omitempty" bson:"username,omitempty"`
	StockSymbol    string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	Filename       string `json:"filename,omitempty" bson:"filename,omitempty"`
	Funds          *int   `json:"funds,omitempty" bson:"funds,omitempty"`
	ErrorMessage   string `json:"errorMessage,omitempty" bson:"errorMessage,omitempty"`
}

// DebugEventInfo audit message for any system debug events
type DebugEventInfo struct {
	LogType        string `json:"logType" bson:"logType"`
	Timestamp      int32  `json:"timestamp" bson:"timestamp"`
	Server         string `json:"server" bson:"server"`
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Command        string `json:"command" bson:"command"`
	Username       string `json:"username,omitempty" bson:"username,omitempty"`
	StockSymbol    string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	Filename       string `json:"filename,omitempty" bson:"filename,omitempty"`
	Funds          *int   `json:"funds,omitempty" bson:"funds,omitempty"`
	DebugMessage   string `json:"debugMessage,omitempty" bson:"debugMessage,omitempty"`
}
