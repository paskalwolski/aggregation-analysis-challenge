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

	//Could follow the tooltip and use a NewTimer + Stop(), but timer efficiency is not the biggest concern here
	boom := time.After(3 * time.Second)
	//Flag Channel for reading response
	scan := make(chan bool)
	//Create a buffered reader for the SSE response
	scanner := bufio.NewScanner(resp.Body)

	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		// if scanner.Err() != nil {
		// TODO: Handle errors with the response scanning
		// }
		// Signal that channel read is complete if EOF or Error reached
		scan <- true
	}()

ReadLoop:
	for {
		select {
		case <-scan:
			fmt.Println("Scanner Closed: Internal")
			break ReadLoop

		case stamp := <-boom:
			fmt.Printf("Time Channel Closed: %v", stamp)
			resp.Body.Close()
		}
	}

	fmt.Printf("Requested for %v\n", time.Since(startTime))
	return
}
