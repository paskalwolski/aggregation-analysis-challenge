package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type subscriber struct {
	duration time.Duration
	key      string
	value    int
	channel  chan stats
}
type publisher struct {
	subscribers  []subscriber
	targetStream *http.Response
}

func main() {
	fmt.Println("Starting the Aggregation Server")
	// Creating the publisher
	pub := publisher{
		make([]subscriber, 0),
		getSSEResponse(),
	}

	http.HandleFunc("/analysis", func(w http.ResponseWriter, r *http.Request) {
		go handleSubscriberWithResource(w, r, &pub)
	})
	http.ListenAndServe(":8080", nil)
}

func handleSubscriberWithResource(w http.ResponseWriter, r *http.Request, pub *publisher) {
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
		dur, err := time.ParseDuration(duration)

		sub := newSub(dur, dimension)
		pub.addSubscriber(sub)
		boom := time.After(dur)
		select {
		case <-boom:
			fmt.Printf("Time Channel Closed")
			pub.unSub(&sub)
		}

	} else {
		http.Error(w, "method not supported", 404)
	}
}

func (p *publisher) addSubscriber(s subscriber) {
	p.subscribers = append(p.subscribers, s)
}

func (p *publisher) unSub(s *subscriber){
	// remove the sub  from the pub's sub list
}

func newSub(dur time.Duration, dim string) (s subscriber) {
	return subscriber{
		dur,
		dim,
		0,
		make(chan stats),
	}
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
