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
			http.Error(w, "Error Parsing URL Queries", 500)
			return
		}
		dimension := q.Get("dimension")
		duration := q.Get("duration")
		if dimension == "" || duration == "" {
			http.Error(w, "Invalid Query Provided", 400)
			return
		}
		res := HandleQuery(duration, dimension)
		jsonRes, err := json.Marshal(res)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding JSON data: %v", err), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Method not supported"))
	}
}
