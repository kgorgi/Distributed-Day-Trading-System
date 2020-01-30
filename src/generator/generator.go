package main

import (
	"bufio"
	user "extremeWorkload.com/daytrader/lib/user"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func parseLine(line string) (int, []string) {
	// Expects line of form "[1] ADD,userid,100"
	spaceSplit := strings.Split(line, " ")
	lineNumber, _ := strconv.Atoi(strings.Trim(spaceSplit[0], "[]"))
	commaSplit := strings.Split(spaceSplit[1], ",")

	return lineNumber, commaSplit
}

func handleLine(line string) (int, error) {
	var status int
	var body string
	var err error

	lineNumber, command := parseLine(line)
	switch command[0] {
	case "ADD":
		status, body, err = user.AddRequest(command[1], command[2])
	case "QUOTE":
		status, body, err = user.QuoteRequest(command[1], command[2])
	case "BUY":
		status, body, err = user.BuyRequest(command[1], command[2], command[3])
	case "COMMIT_BUY":
		status, body, err = user.CommitBuyRequest(command[1])
	case "CANCEL_BUY":
		status, body, err = user.CancelBuyRequest(command[1])
	case "SELL":
		status, body, err = user.SellRequest(command[1], command[2], command[3])
	case "COMMIT_SELL":
		status, body, err = user.CommitSellRequest(command[1])
	case "CANCEL_SELL":
		status, body, err = user.CancelSellRequest(command[1])
	case "SET_BUY_AMOUNT":
		status, body, err = user.SetBuyAmountRequest(command[1], command[2], command[3])
	case "CANCEL_SET_BUY":
		status, body, err = user.CancelSetBuyRequest(command[1], command[2])
	case "SET_BUY_TRIGGER":
		status, body, err = user.SetBuyTriggerRequest(command[1], command[2], command[3])
	case "SET_SELL_AMOUNT":
		status, body, err = user.SetSellAmountRequest(command[1], command[2], command[3])
	case "CANCEL_SET_SELL":
		status, body, err = user.CancelSetSellRequest(command[1], command[2])
	case "SET_SELL_TRIGGER":
		status, body, err = user.SetSellTriggerRequest(command[1], command[2], command[3])
	case "DUMPLOG":
		if len(command) > 2 {
			status, body, err = user.DumplogRequest(command[1], command[2])
		} else {
			status, body, err = user.DumplogRequest("", command[1])
		}
	case "DISPLAY_SUMMARY":
		status, body, err = user.DisplaySummaryRequest(command[1])
	}
	if err != nil {
		return lineNumber, err
	}
	fmt.Println(line + " " + strconv.Itoa(status) + " - " + body)
	return lineNumber, nil
}

func main() {
	var workloadFilePath string
	flag.StringVar(&workloadFilePath, "file", "./workload.txt", "path for workload file")
	flag.Parse()

	fmt.Println("Opening file: " + workloadFilePath)
	file, err := os.Open(workloadFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)
	var readErr error
	var line string
	var lineNumber int
	for {
		line, readErr = reader.ReadString('\n')
		if readErr != nil {
			break
		}

		lineNumber, err = handleLine(line)
		if err != nil {
			fmt.Println(err.Error() + "\nFailed on line" + strconv.Itoa(lineNumber))
			return
		}
		if lineNumber > 2 {
			break
		}

	}

	if readErr != nil && readErr != io.EOF {
		fmt.Println(err)
		fmt.Printf(" > Failed!: %v\n", err)
	}
	fmt.Println("Finished workload generation")

}
