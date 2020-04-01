package main

import (
	"fmt"
	"os"
	"time"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
	"extremeWorkload.com/daytrader/lib/serverurls"
)

type healthChecker func(string) (string, error)

const waitInterval = 5

var sslCertLocation string

func checkHelper(watchUrls map[string][]string, check healthChecker, servertype string) {

	for _, url := range watchUrls[servertype] {
		_, err := check(url)

		if err != nil {
			lib.Error("Bad - %s(%s) - Bad \n%s\n\n", servertype, url, err.Error())
		} else {
			fmt.Printf("Good - %s(%s) \n", servertype, url)
		}
	}
}

func countActiveTriggers(urls []string) int {
	triggersActive := 0
	for _, url := range urls {
		_, message, err := lib.ClientSendRequest(url, lib.HealthCheck)
		if err == nil && message == "ACTIVE" {
			triggersActive++
		}
	}
	return triggersActive
}

func triggerWatch(watchUrls map[string][]string) {
	transactionUrls := watchUrls["transaction"]

	for {
		fmt.Println("Checking triggers")
		triggersActive := countActiveTriggers(transactionUrls)
		if triggersActive == 1 {
			time.Sleep(waitInterval * time.Second)
			continue
		}
		lib.Error("%d trigger servers found. Counting again\n", triggersActive)
		time.Sleep(5 * time.Second)
		triggersActive = countActiveTriggers(transactionUrls)
		if triggersActive == 1 {
			time.Sleep(waitInterval * time.Second)
			continue
		}

		lib.Error("%d trigger servers found.\n", triggersActive)

		triggerActivated := false
		for !triggerActivated {
			lib.Errorln("Attempting to start a trigger sever...")

			for _, url := range transactionUrls {
				_, message, err := lib.ClientSendRequest(url, lib.HealthCheck+"|START")
				if err == nil && message == "ACTIVE" {
					fmt.Println("Trigger server activated")
					triggerActivated = true
					break
				}
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	sslCertLocation = os.Getenv("CLIENT_SSL_CERT_LOCATION")
	security.InitCryptoKey()
	watchUrls := serverurls.GetUrlsConfig().Watch

	lib.DebuggingEnabled = false

	fmt.Println("Starting trigger watch")
	go triggerWatch(watchUrls)

	fmt.Println("Starting Watch")
	for {
		checkHelper(watchUrls, TCPHealthCheck, "transaction")
		checkHelper(watchUrls, TCPHealthCheck, "quote-cache")
		checkHelper(watchUrls, TCPHealthCheck, "audit")
		checkHelper(watchUrls, TCPHealthCheck, "transaction-load")
		checkHelper(watchUrls, HTTPHealthCheck, "web")
		checkHelper(watchUrls, HTTPHealthCheck, "web-load")
		checkHelper(watchUrls, MongoHealthCheck, "dbs")
		time.Sleep(waitInterval * time.Second)
	}
}
