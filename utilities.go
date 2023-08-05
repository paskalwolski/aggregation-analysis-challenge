package main

import (
	"bufio"
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

	fmt.Printf("Dim: %v\tDur: %v\n", dim, dur)

	client := &http.Client{
		// Timeout: time.Second * 1,
	}

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
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan(){
		fmt.Println(scanner.Text())
	}

	return
}
