package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"extremeWorkload.com/daytrader/lib"
	"extremeWorkload.com/daytrader/lib/user"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TCPHealthCheck send a health check to a tcp server
func TCPHealthCheck(url string) (string, error) {
	status, message, err := lib.ClientSendRequest(url, lib.HealthCheck)
	if err != nil {
		return message, err
	}
	if status != lib.StatusOk {
		return message, fmt.Errorf("Expected 200, got %d", status)
	}
	return message, nil
}

// HTTPHealthCheck send a health check to an http server
func HTTPHealthCheck(url string) (string, error) {
	client, err := user.CreateClient("https://"+url+"/", sslCertLocation)
	status, message, err := client.HeartRequest()
	if err != nil {
		return "", err
	}
	if status != lib.StatusOk {
		return message, errors.New("Status not ok: " + strconv.Itoa(status))
	}
	return message, nil
}

// MongoHealthCheck send a health check to a mongo server
func MongoHealthCheck(url string) (string, error) {
	clientOptions := options.Client().ApplyURI(url)
	var err error
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return "", err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		client.Disconnect(context.TODO())
		return "", err
	}

	client.Disconnect(context.TODO())
	return "", nil
}
