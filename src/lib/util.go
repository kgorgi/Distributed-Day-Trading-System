package lib

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var env = os.Getenv("ENV")

// DebuggingEnabled returns true if debugging should be on
var DebuggingEnabled = env == "" || env == "DEV" || env == "DEV-LAB"

// IsLab returns true if in the lab environment
var IsLab = env == "LAB" || env == "DEV-LAB"

// DollarsToCents converts a dollar string to uint cents
func DollarsToCents(dollars string) uint64 {
	data := strings.Split(dollars, ".")
	dollarsNum, _ := strconv.ParseUint(data[0], 10, 64)
	centsNum, _ := strconv.ParseUint(data[1], 10, 64)

	return dollarsNum*uint64(100) + centsNum
}

// CentsToDollars converts uint cents to a dollar string
func CentsToDollars(cents uint64) string {
	dollars := cents / uint64(100)
	remainingCents := cents % uint64(100)
	return fmt.Sprintf("%d.%02d", dollars, remainingCents)
}

// Debugln only prints to console if environment variable is empty or DEV
func Debugln(msg string) {
	if DebuggingEnabled {
		fmt.Println(msg)
	}
}

// GetUnixTimestamp gets the unix time in milliseconds of the server
func GetUnixTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}
