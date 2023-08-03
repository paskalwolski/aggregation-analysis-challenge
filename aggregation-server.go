package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	fmt.Println("Starting the Aggregation Server")
	http.HandleFunc("/analysis", handleGetAnalysis)

	http.ListenAndServe(":8080", nil)
}

func handleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Error Parsing URL Queries", 500)
			return
		}
		dimension := q.Get("dimension")
		duration := q.Get("duration")
		if dimension == "" || duration == "" {
			http.Error(w, "Invalid Query Provided", 400)
			return
		}

		fmt.Printf("Dim: %v\tDur: %v", dimension, duration)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Method not supported"))
	}
}
