package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting the Aggregation Server")

	http.HandleFunc("/analysis", handleGetAnalysis)

	http.ListenAndServe(":8080", nil)
}

func handleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("Success"))

	} else {
		w.WriteHeader(405)
		w.Write([]byte("Method not supported"))
	}
}
