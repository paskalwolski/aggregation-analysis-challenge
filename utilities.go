package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AnalysisResponse struct {
	TotalPosts       int     `json:"total_posts"`
	MinimumTimestamp int64   `json:"minimum_timestamp"`
	MaximumTimestamp int64   `json:"maximum_timestamp"`
	AverageDimension float64 `json:"average_dimension"`
}

var minTime, maxTime time.Time

func HandleAnalysisQuery(dur, dim string) (analysisResponse AnalysisResponse, aError error) {
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
	// Create counters for analysis
	var dimensionCounter, dataCounter float64 = 0, 0

ScanLoop:
	for scanner.Scan() {
		select {
		case <-boom:
			fmt.Printf("Boom Took %v\n", time.Since(startTime))
			fmt.Printf("Time Channel Closed\n")
			break ScanLoop
		default:
			text := scanner.Text()
			// Catch empty lines between messages
			if text != "" {
				dataCounter++
				// Append opening { to json line
				trimmedText := fmt.Sprintf("%v", text[6:])
				var root map[string]map[string]any
				// store the {post_type: data} object
				err = json.Unmarshal([]byte(trimmedText), &root)
				if err != nil {
					fmt.Println(err)
				}
				// Check for the 'data' key
				for k, v := range root {
					// default numeric value from extracting map[]any is float64
					// Extract Timestamp as default float value
					if timestampData, exists := v["timestamp"]; exists {
						// Converting dimension data to explicit float64. This has to be done first?
						var floatTimestamp float64
						var ok bool
						floatTimestamp, ok = timestampData.(float64)
						if !ok {
							fmt.Println("Error converting timestamp data")
						} else {
							handleTimeCheck(floatTimestamp)
						}
					}
					// See note above about extracting value from map[]any
					if dimData, exists := v[dim]; exists {
						fmt.Printf("Found data %v of %T: %v\n", dim, dimData, dimData)
						var dimFloatVal float64
						var ok bool
						// Could switch dimension, and then assert accordingly - but only asserting to float64 for now.
						dimFloatVal, ok = dimData.(float64)
						if !ok {
							fmt.Println("Error converting to float64")
						}
						dimensionCounter += dimFloatVal
					} else {
						fmt.Printf("Key Not Found: %v on type %v\n", dim, k)
					}
				}
			}
		}
	}
	if scanner.Err() != nil {
		aError = scanner.Err()
		return
	}
	fmt.Printf("Reading Request for %v\n", time.Since(startTime))

	analysisResponse.TotalPosts = int(dataCounter)
	analysisResponse.AverageDimension = dimensionCounter / dataCounter
	analysisResponse.MaximumTimestamp = maxTime.Unix()
	analysisResponse.MinimumTimestamp = minTime.Unix()

	return
}

func handleTimeCheck(t float64) {
	unixT := time.Unix(int64(t), 0)
	if minTime.IsZero() {
		minTime = unixT
	} else if unixT.Before(minTime) {
		minTime = unixT
	}
	if maxTime.Before(unixT) {
		maxTime = unixT
	}
}
