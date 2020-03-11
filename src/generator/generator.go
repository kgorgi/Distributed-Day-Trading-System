package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	user "extremeWorkload.com/daytrader/lib/user"
)

var transactionCount uint64 = 0

func handleCommand(command []string) error {
	var status int
	var body string
	var err error = nil

	switch command[1] {
	case "ADD":
		status, body, err = user.AddRequest(command[2], command[3])
	case "QUOTE":
		status, body, err = user.QuoteRequest(command[2], command[3])
	case "BUY":
		status, body, err = user.BuyRequest(command[2], command[3], command[4])
	case "COMMIT_BUY":
		status, body, err = user.CommitBuyRequest(command[2])
	case "CANCEL_BUY":
		status, body, err = user.CancelBuyRequest(command[2])
	case "SELL":
		status, body, err = user.SellRequest(command[2], command[3], command[4])
	case "COMMIT_SELL":
		status, body, err = user.CommitSellRequest(command[2])
	case "CANCEL_SELL":
		status, body, err = user.CancelSellRequest(command[2])
	case "SET_BUY_AMOUNT":
		status, body, err = user.SetBuyAmountRequest(command[2], command[3], command[4])
	case "CANCEL_SET_BUY":
		status, body, err = user.CancelSetBuyRequest(command[2], command[3])
	case "SET_BUY_TRIGGER":
		status, body, err = user.SetBuyTriggerRequest(command[2], command[3], command[4])
	case "SET_SELL_AMOUNT":
		status, body, err = user.SetSellAmountRequest(command[2], command[3], command[4])
	case "CANCEL_SET_SELL":
		status, body, err = user.CancelSetSellRequest(command[2], command[3])
	case "SET_SELL_TRIGGER":
		status, body, err = user.SetSellTriggerRequest(command[2], command[3], command[4])
	case "DISPLAY_SUMMARY":
		status, body, err = user.DisplaySummaryRequest(command[2])
	case "DUMPLOG":
		if len(command) > 3 {
			status, body, err = user.DumplogRequest(command[2], command[3])
			if err == nil && status == 200 {
				err = user.SaveDumplog(body, command[3])
			}
		} else {
			status, body, err = user.DumplogRequest("", command[2])
			if err == nil && status == 200 {
				err = user.SaveDumplog(body, command[2])
			}
		}
	case "HEART":
		status, body, err = user.HeartRequest()
	}

	return err
}

func handleUser(userid string, commands [][]string, wg *sync.WaitGroup) {
	for _, command := range commands {
		err := handleCommand(command)
		if err != nil {
			fmt.Println("Failed on user " + userid + " on line " + command[0] + ": " + err.Error())
			os.Exit(1)
			return
		}

		atomic.AddUint64(&transactionCount, 1)
	}

	wg.Done()
}

func loadFile(filepath string) []string {
	fmt.Println("Opening file: " + filepath)
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	file.Close()

	fmt.Println("File loaded")
	return lines
}

func sortByUser(lines []string) map[string][][]string {
	fmt.Println("Sorting by User")
	var byUser = make(map[string][][]string)

	for _, line := range lines {
		command := parseLine(line)
		byUser[command[2]] = append(byUser[command[2]], command)
	}

	fmt.Println("Commands sorted by user")
	return byUser
}

func parseLine(line string) []string {
	// Expects line of form "[1] ADD,userid,100"
	spaceSplit := strings.Split(line, " ")
	lineNumber := strings.Trim(spaceSplit[0], "[]")
	commaSplit := strings.Split(spaceSplit[1], ",")

	return append([]string{lineNumber}, commaSplit...)
}

func createWorkloadFile(numberCommands int, numberUsers int, filename string) error{

	perUserCommands := numberCommands/numberUsers

	fmt.Printf("Creating heartbeat workload file with %d users and %d commmands\n",numberUsers,numberUsers * perUserCommands)

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return err
	}

	commandNum := 1
	for userIdx := 0; userIdx < numberUsers; userIdx++ {
		for i := 0; i < perUserCommands; i++ {
			fmt.Fprintf(f, "[%d] HEART,goh%d\n", commandNum, userIdx)
			commandNum++
		}
	}
	fmt.Fprintf(f,"[%d] DUMPLOG,./testLOG\n",commandNum)
	return nil
}

func main() {

	var filePath string

	flag.StringVar(&filePath, "f", "workload.txt", "filepath of workload")
	makeWorkload := flag.Bool("w", false, "Switch for generating workload")
	U := flag.Int("U", 50, "Number of users")
	N := flag.Int("N", 5000, "Number of commands")
	flag.Parse()

	if *makeWorkload {
		filePath = "heartworkload.txt"
		err := createWorkloadFile(*N, *U, filePath)
		if err != nil{
			fmt.Println(err.Error())
			return 
		}
	}

	user.InitClient()

	lines := loadFile(filePath)

	dumpLogLineNum := uint64(len(lines) - 1)
	userLines := lines[:dumpLogLineNum]

	commandsByUser := sortByUser(userLines)
	numOfUsers := len(commandsByUser)

	fmt.Println("Starting " + strconv.Itoa(numOfUsers) + " goroutines")
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(numOfUsers)
	for userid, commands := range commandsByUser {
		go handleUser(userid, commands, &wg)
	}

	var currentCount = atomic.LoadUint64(&transactionCount)
	for currentCount < dumpLogLineNum {
		fmt.Println("Transaction Count: " + strconv.FormatUint(currentCount, 10))
		time.Sleep(10 * time.Second)
		currentCount = atomic.LoadUint64(&transactionCount)
	}

	fmt.Println("Waiting for gorountines to finish")
	wg.Wait()

	elapsed := time.Now().Sub(start)
	fmt.Printf("Elapsed Time %f\n", elapsed.Seconds())
	fmt.Println("Executing DUMPLOG")
	dumpLogCommand := parseLine(lines[dumpLogLineNum])
	handleCommand(dumpLogCommand)

	fmt.Println("Finished workload generation")
}
