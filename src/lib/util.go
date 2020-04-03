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

// PerfLoggingEnabled should the system perform performacne logging
var PerfLoggingEnabled = true

// IsLab returns true if in the lab environment
var IsLab = env == "LAB" || env == "DEV-LAB"

// DollarsToCents converts a dollar string to uint cents
func DollarsToCents(dollars string) (uint64, error) {
	data := strings.Split(dollars, ".")
	dollarsNum, err := strconv.ParseUint(data[0], 10, 64)
	if err != nil {
		return 0, err
	}

	centsNum, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		return 0, err
	}

	results := dollarsNum*uint64(100) + centsNum
	return results, nil
}

// CentsToDollars converts uint cents to a dollar string
func CentsToDollars(cents uint64) string {
	dollars := cents / uint64(100)
	remainingCents := cents % uint64(100)
	return fmt.Sprintf("%d.%02d", dollars, remainingCents)
}

// Error writes printf style to StdError
func Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// Errorln writes a message to StdError
func Errorln(msg string) {
	Error("%s\n", msg)
}

// Debug only prints to console if environment variable is empty or DEV
func Debug(format string, a ...interface{}) {
	if DebuggingEnabled {
		fmt.Printf(format, a...)
	}
}

// Debugln only prints to console if environment variable is empty or DEV
func Debugln(msg string) {
	Debug("%s\n", msg)
}

// GetUnixTimestamp gets the unix time in milliseconds of the server
func GetUnixTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}
