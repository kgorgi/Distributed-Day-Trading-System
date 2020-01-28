package auditclient

// InternalLogInfo is not be used directly. Only by the audit
// client or audit server
type InternalLogInfo struct {
	LogType   string `json:"logType" bson:"logType" `
	Timestamp int64  `json:"timestamp" bson:"timestamp" xml:"timestamp"`
	Server    string `json:"server" bson:"server" xml:"server"`
}

// UserCommandInfo audit message for user commands
type UserCommandInfo struct {
	TransactionNum      int    `json:"transactionNum" bson:"transactionNum"`
	Command             string `json:"command" bson:"command"`
	OptionalUsername    string `json:"username,omitempty" bson:"username"`
	OptionalStockSymbol string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename    string `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFunds       *int   `json:"funds,omitempty" bson:"funds,omitempty"`
}

// QuoteServerResponseInfo audit message for quote server responses
type QuoteServerResponseInfo struct {
	TransactionNum  int    `json:"transactionNum" bson:"transactionNum"`
	Price           int    `json:"price" bson:"price"`
	StockSymbol     string `json:"stockSymbol" bson:"stockSymbol"`
	Username        string `json:"username" bson:"username"`
	QuoteServerTime int    `json:"quoteServerTime" bson:"quoteServerTime"`
	CryptoKey       string `json:"cryptoKey" bson:"cryptoKey"`
}

// AccountTransactionInfo audit message for account transactions
type AccountTransactionInfo struct {
	TransactionNum int    `json:"transactionNum" bson:"transactionNum"`
	Action         string `json:"action" bson:"action"`
	Username       string `json:"username" bson:"username"`
	Funds          int    `json:"funds" bson:"funds"`
}

// SystemEventInfo audit message for any system events
type SystemEventInfo struct {
	TransactionNum      int    `json:"transactionNum" bson:"transactionNum"`
	Command             string `json:"command" bson:"command"`
	OptionalUsername    string `json:"username,omitempty" bson:"username,omitempty"`
	OptionalStockSymbol string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename    string `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFunds       *int   `json:"funds,omitempty" bson:"funds,omitempty"`
}

// ErrorEventInfo audit message for any system error events
type ErrorEventInfo struct {
	TransactionNum       int    `json:"transactionNum" bson:"transactionNum"`
	Command              string `json:"command" bson:"command"`
	OptionalUsername     string `json:"username,omitempty" bson:"username,omitempty"`
	OptionalStockSymbol  string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFunds        *int   `json:"funds,omitempty" bson:"funds,omitempty"`
	OptionalErrorMessage string `json:"errorMessage,omitempty" bson:"errorMessage,omitempty"`
}

// DebugEventInfo audit message for any system debug events
type DebugEventInfo struct {
	TransactionNum       int    `json:"transactionNum" bson:"transactionNum"`
	Command              string `json:"command" bson:"command"`
	OptionalUsername     string `json:"username,omitempty" bson:"username,omitempty"`
	OptionalStockSymbol  string `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFunds        *int   `json:"funds,omitempty" bson:"funds,omitempty"`
	OptionalDebugMessage string `json:"debugMessage,omitempty" bson:"debugMessage,omitempty"`
}
