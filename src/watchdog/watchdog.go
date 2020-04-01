package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/security"
	"extremeWorkload.com/daytrader/lib/serverurls"
	"extremeWorkload.com/daytrader/lib/user"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type healthChecker func(string) error

const timeInterval = 60

var sslCertLocation string

// TCPHealthCheck send a health check to a tcp server
func TCPHealthCheck(url string) error {
	status, _, err := lib.ClientSendRequest(url, lib.HealthCheck)
	if err != nil {
		return err
	}
	if status != lib.StatusOk {
		return fmt.Errorf("Expected 200, got %d", status)
	}
	return nil
}

// HTTPHealthCheck send a health check to an http server
func HTTPHealthCheck(url string) error {
	client, err := user.CreateClient("https://"+url+"/", sslCertLocation)
	status, _, err := client.HeartRequest()
	if err != nil {
		return err
	}
	if status != lib.StatusOk {
		return errors.New("Status not ok: " + strconv.Itoa(status))
	}
	return nil
}

// MongoHealthCheck send a health check to a mongo server
func MongoHealthCheck(url string) error {
	clientOptions := options.Client().ApplyURI(url)
	var err error
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		client.Disconnect(context.TODO())
		return err
	}

	client.Disconnect(context.TODO())
	return nil
}

func checkHelper(watchUrls map[string][]string, check healthChecker, servertype string) {
	fmt.Printf("--%s--\n", servertype)
	for _, url := range watchUrls[servertype] {
		fmt.Printf("Checking %s... ", url)
		err := check(url)
		if err != nil {
			fmt.Printf("Bad \n%s\n", err.Error())
		} else {
			fmt.Println("Good")
		}
	}
}

func main() {
	sslCertLocation = os.Getenv("CLIENT_SSL_CERT_LOCATION")
	security.InitCryptoKey()
	watchUrls := serverurls.GetUrlsConfig().Watch

	fmt.Println("Starting Watch")
	for {
		checkHelper(watchUrls, TCPHealthCheck, "transaction")
		checkHelper(watchUrls, TCPHealthCheck, "quote-cache")
		checkHelper(watchUrls, TCPHealthCheck, "audit")
		checkHelper(watchUrls, TCPHealthCheck, "transaction-load")
		checkHelper(watchUrls, HTTPHealthCheck, "web")
		checkHelper(watchUrls, HTTPHealthCheck, "web-load")
		checkHelper(watchUrls, MongoHealthCheck, "dbs")
		fmt.Printf("Round complete. waiting\n\n")
		time.Sleep(timeInterval * time.Second)
	}
}
