package main

import "errors"

var (
	// ErrDataNotFound is returned when data from the data server can't be found
	ErrDataNotFound = errors.New("data not found")
)
