package main

import (
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

const userID = "username"
const stockSymbol = "stockSymbol"
const filename = "filename"
const fundsInCents = "funds"

func writeInternalLogInfoTags(
	str *strings.Builder,
	internalInfo auditclient.InternalLogInfo,
	logCommand bool) {
	writeStringTag(str, "timestamp", strconv.FormatUint(internalInfo.Timestamp, 10))
	writeStringTag(str, "server", internalInfo.Server)
	writeStringTag(str, "transactionNum", strconv.FormatUint(internalInfo.TransactionNum, 10))

	if logCommand && internalInfo.Command != "" {
		writeStringTag(str, "command", internalInfo.Command)
	}
}

func writeUserCommandTags(str *strings.Builder, info auditclient.UserCommandInfo) {
	writeOptionalStringTag(str, userID, info.OptionalUserID)
	writeOptionalStringTag(str, stockSymbol, info.OptionalStockSymbol)
	writeOptionalStringTag(str, filename, info.OptionalFilename)
	writeOptionalDecimalTag(str, fundsInCents, info.OptionalFundsInCents)
}

func writeQuoteServerTags(str *strings.Builder, info auditclient.QuoteServerResponseInfo) {
	writeDecimalTag(str, "price", info.PriceInCents)
	writeStringTag(str, stockSymbol, info.StockSymbol)
	writeStringTag(str, userID, info.UserID)
	writeStringTag(str, "quoteServerTime", strconv.FormatUint(info.QuoteServerTime, 10))
	writeStringTag(str, "cryptokey", info.CryptoKey)
}

func writeAccountTransactionTags(str *strings.Builder, info auditclient.AccountTransactionInfo) {
	writeStringTag(str, "action", info.Action)
	writeStringTag(str, userID, info.UserID)
	writeDecimalTag(str, fundsInCents, info.FundsInCents)
}

func writeErrorEventTags(str *strings.Builder, info auditclient.ErrorEventInfo) {
	writeStringTag(str, "errorMessage", info.ErrorMessage)
}

func writeDebugEventTags(str *strings.Builder, info auditclient.DebugEventInfo) {
	writeStringTag(str, "debugMessage", info.DebugMessage)
}

func writePerfMetricTags(str *strings.Builder, info auditclient.PerformanceMetricInfo) {
	writeStringTag(str, "acceptTimestamp", strconv.FormatUint(info.AcceptTimestamp, 10))
	writeStringTag(str, "readTimestamp", strconv.FormatUint(info.ReadTimestamp, 10))
	writeStringTag(str, "writeTimestamp", strconv.FormatUint(info.WriteTimestamp, 10))
	writeStringTag(str, "closeTimestamp", strconv.FormatUint(info.CloseTimestamp, 10))
}

func writeOptionalDecimalTag(str *strings.Builder, tag string, amount *uint64) {
	if amount != nil {
		writeDecimalTag(str, tag, *amount)
	}
}

func writeDecimalTag(str *strings.Builder, tag string, amount uint64) {
	str.WriteString("\t\t")
	writeTagHead(str, tag)
	str.WriteString(lib.CentsToDollars(amount))
	writeTagTail(str, tag)
	str.WriteString("\n")
}

func writeOptionalStringTag(str *strings.Builder, tag string, info string) {
	if info != "" {
		writeStringTag(str, tag, info)
	}
}

func writeStringTag(str *strings.Builder, tag string, value string) {
	str.WriteString("\t\t")
	writeTagHead(str, tag)
	str.WriteString(value)
	writeTagTail(str, tag)
	str.WriteString("\n")
}

func writeTagHead(str *strings.Builder, tag string) {
	str.WriteString("<")
	str.WriteString(tag)
	str.WriteString(">")
}

func writeTagTail(str *strings.Builder, tag string) {
	str.WriteString("</")
	str.WriteString(tag)
	str.WriteString(">")
}
