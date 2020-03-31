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

const timeInterval = 60

var sslCertLocation string

func checkHelper(watchUrls map[string][]string, check healthChecker, servertype string) []string {
	var replies []string

	fmt.Printf("%s:\n", servertype)

	for _, url := range watchUrls[servertype] {
		fmt.Printf("Checking %s... ", url)
		reply, err := check(url)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad \n%s\n\n", err.Error())
		} else {
			fmt.Println("Good")
			replies = append(replies, reply)
		}
	}
	return replies
}

func find(query string, list []string) int {
	c := 0
	for _, element := range list {
		if query == element {
			c++
		}
	}
	return c
}

func main() {
	sslCertLocation = os.Getenv("CLIENT_SSL_CERT_LOCATION")
	security.InitCryptoKey()
	watchUrls := serverurls.GetUrlsConfig().Watch

	fmt.Println("Starting Watch")
	for {
		replies := checkHelper(watchUrls, TCPHealthCheck, "transaction")
		triggerServerCount := find(lib.HealthStatusTrigger, replies)
		if triggerServerCount != 1 {
			fmt.Fprintf(os.Stderr, "%d trigger servers found\n", triggerServerCount)
		}
		checkHelper(watchUrls, TCPHealthCheck, "quote-cache")
		checkHelper(watchUrls, TCPHealthCheck, "audit")
		checkHelper(watchUrls, TCPHealthCheck, "transaction-load")
		checkHelper(watchUrls, HTTPHealthCheck, "web")
		checkHelper(watchUrls, HTTPHealthCheck, "web-load")
		checkHelper(watchUrls, MongoHealthCheck, "dbs")
		fmt.Printf("--Round complete. waiting %d seconds --\n\n", timeInterval)
		time.Sleep(timeInterval * time.Second)
	}
}
