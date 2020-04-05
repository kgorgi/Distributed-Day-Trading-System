package e2e

import (
	"fmt"
	"os"
	"testing"

	user "extremeWorkload.com/daytrader/lib/user"
)

var userClient *user.UserClient

const webserverAddress = "https://localhost:8080/"

func TestMain(m *testing.M) {
	var err error
	userClient, err = user.CreateClient(webserverAddress, os.Getenv("CLIENT_SSL_CERT_LOCATION"))
	if err != nil {
		fmt.Println("Failed while creating a user client")
		fmt.Println(err)
		os.Exit(1)
		return
	}
	os.Exit(m.Run())
}
