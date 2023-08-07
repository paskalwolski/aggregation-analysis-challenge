package main

import (
	"bufio"
	"encoding/json"
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
	funcTime := time.Now()
	defer func() { fmt.Printf("Function Execution Took %v\n\n", time.Since(funcTime)) }()
	duration, err := time.ParseDuration(dur)
	if err != nil {
		aError = err
		return
	}

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
	defer func() { fmt.Printf("Request Open for %v\n", time.Since(startTime)) }()
	defer resp.Body.Close()

	// Could follow the tooltip and use a NewTimer + Stop(), but timer efficiency is not the biggest concern here
	boom := time.After(duration)

	// Create a buffered reader for the SSE response
	scanner := bufio.NewScanner(resp.Body)

ScanLoop:
	for scanner.Scan() {
		select {
		case <-boom:
			fmt.Printf("Time Channel Closed\n")
			break ScanLoop
		default:
			text := scanner.Text()
			if text != "" {
				trimmedText := fmt.Sprintf("{%v", text[7:])
				fmt.Println(trimmedText)
				var m map[string]any
				err = json.Unmarshal([]byte(trimmedText), &m)
				if err != nil {
					fmt.Println(err)
				}

			}
		}
	}
	fmt.Printf("Reading Request for %v\n", time.Since(startTime))
	if scanner.Err() != nil {
		aError = scanner.Err()
		return
	}
	return
}
