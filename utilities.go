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
	funcTime := time.Now()
	defer func() { fmt.Printf("Function Execution Took %v", time.Since(funcTime)) }()
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

	//Could follow the tooltip and use a NewTimer + Stop(), but timer efficiency is not the biggest concern here
	boom := time.After(duration)
	//Flag Channel for reading response
	scan := make(chan bool)
	//Create a buffered reader for the SSE response
	scanner := bufio.NewScanner(resp.Body)

	go func(parentError error) {
	ScanLoop:
		for scanner.Scan() {
			select {
			case <-scan:
				//Interrupting the Read Loop externally
				break ScanLoop
			default:
				//Normal Read Operation:
				fmt.Println(scanner.Text())
			}
		}
		if scanner.Err() != nil {
			fmt.Println(scanner.Err())
			aError = scanner.Err()
		}
		// Signal that channel read is complete if EOF or Error reached
		scan <- true
	}(aError)

	select {
	case <-scan:
		fmt.Println("Scanner Closed: Internal")
	case <-boom:
		fmt.Printf("Time Channel Closed\n")
		scan <- true
	}
	if aError != nil {
		return
	}
	fmt.Printf("Requested for %v\n", time.Since(startTime))
	return
}
