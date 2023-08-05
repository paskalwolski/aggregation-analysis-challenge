package main

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

type AnalysisResponse struct {
	Amount           int16
	TimestampStart   string
	TimestampEnd     string
	AverageDimension float32
}

func HandleAnalysisQuery(dur, dim string) (aResponse AnalysisResponse, aError error) {

	// parseDuration

	fmt.Printf("Dim: %v\tDur: %v\n", dim, dur)

	client := &http.Client{}

	sseRequest, err := http.NewRequest(http.MethodGet, "https://stream.upfluence.co/stream", nil)
	if err != nil {
		aError = err
		return
	}

	sseRequest.Header.Add("Accept", "text/event-stream")

	resp, err := client.Do(sseRequest)
	if err != nil {
		aError = err
		return
	}
	startTime := time.Now()
	defer resp.Body.Close()

	boom := time.After(3 * time.Second)
	scan := make(chan bool)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		scan <- true
	}()

ReadLoop:
	for {
		select {
		case <-scan:
			fmt.Println("Scanner Closed")

		case <-boom:
			break ReadLoop
		}
	}

	fmt.Printf("Requested for %v\n", time.Since(startTime))
	return
}
