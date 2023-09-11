package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type subscriber struct {
	id       int64
	duration time.Duration
	key      string
	value    int
	channel  chan string
}
type publisher struct {
	subscribers  []*subscriber
	targetStream *http.Response
	subscribeMu  *sync.Mutex
	subsMu       *sync.RWMutex
}

func main() {
	fmt.Println("Starting the Aggregation Server")
	// Creating the publisher
	pub := publisher{
		make([]*subscriber, 0),
		getSSEResponse(),
		&sync.Mutex{},
		&sync.RWMutex{},
	}
	fmt.Println("New publisher created.")
	go runPub(&pub)

	http.HandleFunc("/analysis", func(w http.ResponseWriter, r *http.Request) {
		handleSubscriberWithResource(w, r, &pub)
	})
	http.ListenAndServe(":8080", nil)
}

func runPub(p *publisher) {
	scanner := bufio.NewScanner(p.targetStream.Body)
	// Create counters for analysis
	// var dimensionCounter, dataCounter float64 = 0, 0
	// Create time trackers
	// var minTime, maxTime time.Time

	for scanner.Scan() {
		t := scanner.Text()
		// fmt.Println(scanner.Text())
		p.subsMu.RLock()
		for _, s := range p.subscribers {
			s.channel <- t
		}
		p.subsMu.RUnlock()
	}
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

		sub := newSub(dur, dimension, pub.subscribeMu)
		pub.addSubscriber(sub)
		go sub.getSubData()
		boom := time.After(dur)
		<-boom
		fmt.Printf("Time Channel Closed\n")
		pub.unSub(sub)

	} else {
		http.Error(w, "method not supported", 404)
	}
}

func (s *subscriber) getSubData() {
	for v := range s.channel{
		fmt.Println(v)
	}
}

func (p *publisher) addSubscriber(s *subscriber) {
	p.subsMu.Lock()
	p.subscribers = append(p.subscribers, s)
	p.subsMu.Unlock()
}

func (p *publisher) unSub(sub *subscriber) {
	// remove the sub  from the pub's sub list
	for i, s := range p.subscribers {
		if s.id == sub.id {
			p.subsMu.Lock()
			fmt.Printf("Ready to unsub at position %v\n", i)
			// Quick 2-step to pop item. Replace index with last element
			p.subscribers[i] = p.subscribers[len(p.subscribers)-1]
			// trim last element
			p.subscribers = p.subscribers[:len(p.subscribers)-1]
			p.subsMu.Unlock()
			close(s.channel)
		}
	}
}

func newSub(dur time.Duration, dim string, mu *sync.Mutex) (s *subscriber) {
	s = &subscriber{
		createNewSubId(mu),
		dur,
		dim,
		0,
		make(chan string),
	}
	fmt.Printf("Created new sub with ID %v\n", s.id)
	return
}

func createNewSubId(mu *sync.Mutex) (id int64) {
	mu.Lock()
	id = time.Now().Unix()
	mu.Unlock()
	return
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
