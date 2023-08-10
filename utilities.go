package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
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
	defer func() { log.Printf("Function Execution Took %v\n\n", time.Since(funcTime)) }()
	duration, err := time.ParseDuration(dur)
	if err != nil {
		aError = fmt.Errorf("error parsing duration")
		log.Println(err)
		return
	}
	fmt.Printf("TIMEPARSE: %v\n", time.Since(funcTime).Seconds())
	sseResponse, responseStartTime, respErr := getSSEResponse()
	if respErr != nil {
		log.Println(respErr)
		aError = fmt.Errorf("error accessing sse stream")
		return
	}
	fmt.Printf("RESPONSEOPEN: %v\n", time.Since(funcTime).Seconds())
	// Check how long the response was open for based on specified duration
	defer func() { log.Printf("Request Open for %v\n", time.Since(responseStartTime)) }()
	defer sseResponse.Body.Close()

	// Could follow the tooltip and use a NewTimer + Stop(), but timer efficiency is not the biggest concern here
	boom := time.After(duration)

	// Create a buffered reader for the SSE response
	scanner := bufio.NewScanner(sseResponse.Body)
	fmt.Printf("SCANOPEN: %v\n", time.Since(funcTime).Seconds())
	// Create counters for analysis
	var dimensionCounter, dataCounter float64 = 0, 0

	// Loop that runs each time there is a new value on the scanner
	// It checks that there is still time left, and if so proceeds to read the value.
	// If the time has run out, the boom case is run, the current value is discarded - and the loop does not run again.
ScanLoop:
	for scanner.Scan() {
		select {
		case <-boom:
			// Time Channel Closed
			break ScanLoop
		default:
			text := scanner.Text()
			// Catch empty lines between messages
			if text != "" {
				dataCounter++
				// Append opening { to json line
				trimmedText := fmt.Sprintf("%v", text[6:])
				// store the {post_type: data} object
				var root map[string]map[string]any
				err = json.Unmarshal([]byte(trimmedText), &root)
				if err != nil {
					log.Println(err)
					aError = fmt.Errorf("error reading stream json data")
					return
				}
				// Check for the 'data' key. We have no idea what this might be, but it has to exist!
				for k, v := range root {
					timestamp, err := handleNumericDataExtraction(v, "timestamp")
					if err != nil {
						aError = fmt.Errorf("error extracting key from response")
						// argument order is strange - but k is SSE data type, and v is its value - the map we are searching
						log.Printf("could not extract key %v from data %v", v, k)
					}
					handleTimeCheck(timestamp)

					dimFloatValue, err := handleNumericDataExtraction(v, dim)
					if err != nil {
						aError = fmt.Errorf("error extracting key from response")
						log.Printf("could not extract key %v from data %v", v, k)
					}
					dimensionCounter += dimFloatValue
				}
			}
		}
	}
	if sError := scanner.Err(); sError != nil {
		log.Println(sError)
		aError = fmt.Errorf("error scanning incoming stream data")
		return
	}

	// Build the final Analysis response
	analysisResponse.TotalPosts = int(dataCounter)
	analysisResponse.AverageDimension = dimensionCounter / dataCounter
	analysisResponse.MaximumTimestamp = maxTime.Unix()
	analysisResponse.MinimumTimestamp = minTime.Unix()

	return
}

// The incoming scanner stream is of an unknown shape - so it comes in as map[string]any
// When extracting a value using key k, it is internally type asserted.
// If the value is numeric, this is cast to float64 - and I have not been able to manually override this.
// This function extracts the value and coaxes it into a discrete float64 value, which can be converted further.
func handleNumericDataExtraction(m map[string]any, k string) (val float64, eError error) {
	if data, exists := m[k]; exists {
		// Converting dimension data to explicit float64. This has to be done first?
		var ok bool
		val, ok = data.(float64)
		// Check if ok - otherwise, val is returned at the bottom
		if !ok {
			log.Printf("Key %v not found", k)
			eError = fmt.Errorf("key not found")
			return
		}
	} else {
		// catch conversion errors
		log.Printf("error converting to float64: %v of %T", data, data)
		eError = fmt.Errorf("error converting numeric data %v of %T", data, data)
		return
	}
	// Nothing failed - return the val!
	return
}

// Handle the outbound request to the SSE Stream
func getSSEResponse() (resp *http.Response, startTime time.Time, respErr error) {
	client := &http.Client{}

	sseRequest, err := http.NewRequest(http.MethodGet, "https://stream.upfluence.co/stream", nil)
	if err != nil {
		respErr = err
		return
	}

	sseRequest.Header.Add("Accept", "text/event-stream")

	resp, respErr = client.Do(sseRequest)
	if respErr != nil {
		return
	}

	startTime = time.Now()
	return
}

// Logic for comparing the current time to the existing min and max times
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
