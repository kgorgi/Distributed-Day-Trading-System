package lib

import (
	"fmt"
	"strconv"
	"strings"
)

func DollarsToCents(dollars string) uint64 {
	data := strings.Split(dollars, ".")
	dollarsNum, _ := strconv.ParseUint(data[0], 10, 64)
	centsNum, _ := strconv.ParseUint(data[1], 10, 64)

	return dollarsNum*uint64(100) + centsNum
}

func CentsToDollars(cents uint64) string {
	dollars := cents / uint64(100)
	remainingCents := cents % uint64(100)
	return fmt.Sprintf("%d.%d", dollars, remainingCents)
}
