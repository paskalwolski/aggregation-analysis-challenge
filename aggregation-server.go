package main

import (
	"encoding/json"
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
			http.Error(w, "invalid query data", 400)
			return
		}
		dimension := q.Get("dimension")
		duration := q.Get("duration")
		if dimension == "" || duration == "" {
			http.Error(w, "invalid query value", 400)
			return
		}
		res, err := HandleAnalysisQuery(duration, dimension)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		jsonRes, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "error creating response", 500)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	} else {
		http.Error(w, "method not supported", 404)
	}
}
