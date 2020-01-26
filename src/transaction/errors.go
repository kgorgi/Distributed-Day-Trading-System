package main

import (
	"fmt"
)

type InvalidData struct {
	What string
}

func (e *InvalidData) Error() string {
	return fmt.Sprintf("%s", e.What)
}
