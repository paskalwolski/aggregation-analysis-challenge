package main

import (
	"fmt"
)

type Response struct {
	Amount           int16
	TimestampStart   string
	TimestampEnd     string
	AverageDimension float32
}

func HandleQuery(dur, dim string) Response {
	fmt.Printf("Dim: %v\tDur: %v\n", dim, dur)
	var res Response
	return res
}
