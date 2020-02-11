package lib

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var env = os.Getenv("ENV")
var debuggingEnabled = env == "" || env == "DEV"

func DollarsToCents(dollars string) uint64 {
	data := strings.Split(dollars, ".")
	dollarsNum, _ := strconv.ParseUint(data[0], 10, 64)
	centsNum, _ := strconv.ParseUint(data[1], 10, 64)

	return dollarsNum*uint64(100) + centsNum
}

func CentsToDollars(cents uint64) string {
	dollars := cents / uint64(100)
	remainingCents := cents % uint64(100)
	return fmt.Sprintf("%d.%02d", dollars, remainingCents)
}

func UseLabQuoteServer() bool {
	env := os.Getenv("ENV")
	return env == "LAB"
}

// Debugln only prints to console if environment variable is empty or DEV
func Debugln(msg string) {
	if debuggingEnabled {
		fmt.Println(msg)
	}
}
