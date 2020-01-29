package main

import (
	"fmt"
	"strconv"
	"strings"

	auditclient "extremeWorkload.com/daytrader/lib/audit"
)

const userID = "userID"
const stockSymbol = "stockSymbol"
const filename = "filename"
const fundsInCents = "fundsInCents"

func writeInternalLogInfoTags(
	str *strings.Builder,
	internalInfo auditclient.InternalLogInfo,
	logCommand bool) {
	writeStringTag(str, "timestamp", strconv.FormatInt(internalInfo.Timestamp, 10))
	writeStringTag(str, "server", internalInfo.Server)
	writeStringTag(str, "transactionNum", string(internalInfo.TransactionNum))

	if logCommand && internalInfo.Command != "" {
		writeStringTag(str, "command", internalInfo.Command)
	}
}

func writeUserCommandTags(str *strings.Builder, info auditclient.UserCommandInfo) {
	writeOptionalStringTag(str, userID, info.OptionalUserID)
	writeOptionalStringTag(str, stockSymbol, info.OptionalStockSymbol)
	writeOptionalStringTag(str, filename, info.OptionalFilename)
	writeOptionalDecimalTag(str, "fundsInCents", info.OptionalFundsInCents)
}

func writeQuoteServerTags(str *strings.Builder, info auditclient.QuoteServerResponseInfo) {
	writeDecimalTag(str, "priceInCents", info.PriceInCents)
	writeStringTag(str, stockSymbol, info.StockSymbol)
	writeStringTag(str, userID, info.UserID)
	writeStringTag(str, "quoteServerTime", string(info.QuoteServerTime))
	writeStringTag(str, "cryptokey", info.CryptoKey)
}

func writeAccountTransactionTags(str *strings.Builder, info auditclient.AccountTransactionInfo) {
	writeStringTag(str, "action", info.Action)
	writeStringTag(str, userID, info.UserID)
	writeDecimalTag(str, fundsInCents, info.FundsInCents)
}

func writeSystemEventTags(str *strings.Builder, info auditclient.SystemEventInfo) {
	writeOptionalStringTag(str, userID, info.OptionalUserID)
	writeOptionalStringTag(str, stockSymbol, info.OptionalStockSymbol)
	writeOptionalStringTag(str, filename, info.OptionalFilename)
	writeOptionalDecimalTag(str, fundsInCents, info.OptionalFundsInCents)
}

func writeErrorEventTags(str *strings.Builder, info auditclient.ErrorEventInfo) {
	writeOptionalStringTag(str, userID, info.OptionalUserID)
	writeOptionalStringTag(str, stockSymbol, info.OptionalStockSymbol)
	writeOptionalStringTag(str, filename, info.OptionalFilename)
	writeOptionalDecimalTag(str, fundsInCents, info.OptionalFundsInCents)
	writeOptionalStringTag(str, "errorMessage", info.OptionalErrorMessage)
}

func writeDebugEventTags(str *strings.Builder, info auditclient.DebugEventInfo) {
	writeOptionalStringTag(str, userID, info.OptionalUserID)
	writeOptionalStringTag(str, stockSymbol, info.OptionalStockSymbol)
	writeOptionalStringTag(str, filename, info.OptionalFilename)
	writeOptionalDecimalTag(str, fundsInCents, info.OptionalFundsInCents)
	writeOptionalStringTag(str, "debugMessage", info.OptionalDebugMessage)
}

func writeOptionalDecimalTag(str *strings.Builder, tag string, amount *uint64) {
	if amount != nil {
		writeDecimalTag(str, tag, *amount)
	}
}

func writeDecimalTag(str *strings.Builder, tag string, amount uint64) {
	str.WriteString("\t\t")
	writeTagHead(str, tag)
	decimal := float64(amount) / float64(100)
	fmt.Fprintf(str, "%.2f", decimal)
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
