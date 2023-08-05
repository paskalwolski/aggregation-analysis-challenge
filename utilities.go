package main

import (
	"fmt"
	"net/http"
)

type AnalysisResponse struct {
	Amount           int16
	TimestampStart   string
	TimestampEnd     string
	AverageDimension float32
}

func HandleQuery(dur, dim string) (aResponse AnalysisResponse, aError error) {
	// var aRes AnalysisResponse

	fmt.Printf("Dim: %v\tDur: %v\n", dim, dur)

	client := http.Client{
		// Timeout: time.Second * 1,
	}
	fmt.Println("Starting Request")
	resp, err := client.Get("https://stream.upfluence.co/stream")
	if err != nil {
		aError = err
		return
	}
	if err != nil {
		aError = err
		return
	}

	fmt.Println("Done with Request")
	defer resp.Body.Close()

	return
}
