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

const waitInterval = 10

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

		for triggersActive != 1 {
			lib.Error("%d trigger servers found. Attempting to fix trigger service...\n", triggersActive)

			for _, url := range transactionUrls {
				if triggersActive < 1 {
					_, message, err := lib.ClientSendRequest(url, lib.HealthCheck+"|START")
					if err == nil && message == "STARTED" {
						fmt.Println("Trigger server activated")
						triggersActive++
						break
					}
				} else if triggersActive > 1 {
					_, message, err := lib.ClientSendRequest(url, lib.HealthCheck+"|STOP")
					if err == nil && message == "STOPPED" {
						fmt.Println("Trigger server stopped")
						triggersActive--
						break
					}
				}

			}
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	sslCertLocation = os.Getenv("CLIENT_SSL_CERT_LOCATION")
	if sslCertLocation == "" {
		sslCertLocation = "./ssl/cert.pem"
	}

	security.InitCryptoKey()
	watchUrls := serverurls.GetUrlsConfig().Watch

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
