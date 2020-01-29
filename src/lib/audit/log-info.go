package auditclient

// InternalLogInfo is not be used directly. Only by the audit
// client or audit server
type InternalLogInfo struct {
	LogType        string `json:"logType" bson:"logType" `
	Timestamp      int64  `json:"timestamp" bson:"timestamp" xml:"timestamp"`
	Server         string `json:"server" bson:"server" xml:"server"`
	TransactionNum uint64 `json:"transactionNum" bson:"transactionNum"`
	Command        string `json:"command,omitempty" bson:"command,omitempty"`
}

// UserCommandInfo audit message for user commands
type UserCommandInfo struct {
	OptionalUserID       string  `json:"userID,omitempty" bson:"userID,omitempty"`
	OptionalStockSymbol  string  `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string  `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFundsInCents *uint64 `json:"fundsInCents,omitempty" bson:"fundsInCents,omitempty"`
}

// QuoteServerResponseInfo audit message for quote server responses
type QuoteServerResponseInfo struct {
	PriceInCents    uint64 `json:"priceInCents" bson:"priceInCents"`
	StockSymbol     string `json:"stockSymbol" bson:"stockSymbol"`
	UserID          string `json:"userID" bson:"userID"`
	QuoteServerTime uint64 `json:"quoteServerTime" bson:"quoteServerTime"`
	CryptoKey       string `json:"cryptoKey" bson:"cryptoKey"`
}

// AccountTransactionInfo audit message for account transactions
type AccountTransactionInfo struct {
	Action       string `json:"action" bson:"action"`
	UserID       string `json:"userID" bson:"userID"`
	FundsInCents uint64 `json:"fundsInCents" bson:"fundsInCents"`
}

// SystemEventInfo audit message for any system events
type SystemEventInfo struct {
	OptionalUserID       string  `json:"userID,omitempty" bson:"userID,omitempty"`
	OptionalStockSymbol  string  `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string  `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFundsInCents *uint64 `json:"fundsInCents,omitempty" bson:"fundsInCents,omitempty"`
}

// ErrorEventInfo audit message for any system error events
type ErrorEventInfo struct {
	OptionalUserID       string  `json:"userID,omitempty" bson:"userID,omitempty"`
	OptionalStockSymbol  string  `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string  `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFundsInCents *uint64 `json:"fundsInCents,omitempty" bson:"fundsInCents,omitempty"`
	OptionalErrorMessage string  `json:"errorMessage,omitempty" bson:"errorMessage,omitempty"`
}

// DebugEventInfo audit message for any system debug events
type DebugEventInfo struct {
	OptionalUserID       string  `json:"userID,omitempty" bson:"userID,omitempty"`
	OptionalStockSymbol  string  `json:"stockSymbol,omitempty" bson:"stockSymbol,omitempty"`
	OptionalFilename     string  `json:"filename,omitempty" bson:"filename,omitempty"`
	OptionalFundsInCents *uint64 `json:"fundsInCents,omitempty" bson:"fundsInCents,omitempty"`
	OptionalDebugMessage string  `json:"debugMessage,omitempty" bson:"debugMessage,omitempty"`
}
