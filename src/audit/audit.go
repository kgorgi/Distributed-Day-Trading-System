package main

import (
	"fmt"

	"extremeWorkload.com/daytrader/lib"
)

func main() {
	fmt.Println("Hello, world.")
	lib.ServerSendOKResponse(nil)
}
