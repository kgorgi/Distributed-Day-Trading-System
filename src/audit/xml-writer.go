package main

import (
	"strconv"
	"strings"

	"extremeWorkload.com/daytrader/lib"
	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

func writeInternalLogInfoTags(str *strings.Builder, internalInfo auditclient.InternalLogInfo) {
	writeStringTag(str, "timestamp", strconv.FormatInt(internalInfo.Timestamp, 10))
	writeStringTag(str, "server", internalInfo.Server)
}

func writeUserCommandTags(str *strings.Builder, info auditclient.UserCommandInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeStringTag(str, "command", info.Command)
	writeOptionalStringTag(str, "username", info.OptionalUsername)
	writeOptionalStringTag(str, "stockSymbol", info.OptionalStockSymbol)
	writeOptionalStringTag(str, "filename", info.OptionalFilename)
	writeOptionalDecimalTag(str, "funds", info.OptionalFunds)
}

func writeQuoteServerTags(str *strings.Builder, info auditclient.QuoteServerResponseInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeDecimalTag(str, "price", info.Price)
	writeStringTag(str, "stockSymbol", info.StockSymbol)
	writeStringTag(str, "username", info.Username)
	writeStringTag(str, "quoteServerTime", string(info.QuoteServerTime))
	writeStringTag(str, "cryptokey", info.CryptoKey)
}

func writeAccountTransactionTags(str *strings.Builder, info auditclient.AccountTransactionInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeStringTag(str, "action", info.Action)
	writeStringTag(str, "username", info.Username)
	writeDecimalTag(str, "funds", info.Funds)
}

func writeSystemEventTags(str *strings.Builder, info auditclient.SystemEventInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeStringTag(str, "command", info.Command)

	writeOptionalStringTag(str, "username", info.OptionalUsername)
	writeOptionalStringTag(str, "stockSymbol", info.OptionalStockSymbol)
	writeOptionalStringTag(str, "filename", info.OptionalFilename)
	writeOptionalDecimalTag(str, "funds", info.OptionalFunds)
}

func writeErrorEventTags(str *strings.Builder, info auditclient.ErrorEventInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeStringTag(str, "command", info.Command)

	writeOptionalStringTag(str, "username", info.OptionalUsername)
	writeOptionalStringTag(str, "stockSymbol", info.OptionalStockSymbol)
	writeOptionalStringTag(str, "filename", info.OptionalFilename)
	writeOptionalDecimalTag(str, "funds", info.OptionalFunds)
	writeOptionalStringTag(str, "errorMessage", info.OptionalErrorMessage)
}

func writeDebugEventTags(str *strings.Builder, info auditclient.DebugEventInfo) {
	writeStringTag(str, "transactionNum", string(info.TransactionNum))
	writeStringTag(str, "command", info.Command)

	writeOptionalStringTag(str, "username", info.OptionalUsername)
	writeOptionalStringTag(str, "stockSymbol", info.OptionalStockSymbol)
	writeOptionalStringTag(str, "filename", info.OptionalFilename)
	writeOptionalDecimalTag(str, "funds", info.OptionalFunds)
	writeOptionalStringTag(str, "debugMessage", info.OptionalDebugMessage)
}

func writeOptionalDecimalTag(str *strings.Builder, tag string, amount *int) {
	if amount != nil {
		writeDecimalTag(str, tag, *amount)
	}
}

func writeDecimalTag(str *strings.Builder, tag string, amount int) {
	str.WriteString("\t\t")
	writeTagHead(str, tag)
	str.WriteString(lib.CentsToDollars(uint64(amount)))
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
