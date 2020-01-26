package main

import (
	"fmt"
)

// InvalidData unexpected value for data
type InvalidData struct {
	What string
}

func (e *InvalidData) Error() string {
	return fmt.Sprintf("%s", e.What)
}
