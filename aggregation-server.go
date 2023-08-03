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


func handleGetAnalysis(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Success"))
}
