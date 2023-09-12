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
	channel  chan map[string]interface{}
	open     bool
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

// Publisher main thread.
//
// This interprets the incoming SSE stream, and parses data events into a map.
// The map is sent to the list of subscribers.
func runPub(p *publisher) {
	scanner := bufio.NewScanner(p.targetStream.Body)
	// Create counters for analysis
	// var dimensionCounter, dataCounter float64 = 0, 0
	// Create time trackers
	// var minTime, maxTime time.Time

	for scanner.Scan() {
		t := scanner.Bytes()
		event, body := identifyStreamData(t)
		switch event {
		case "data":
			data := decomposeDataEvent(body)
			p.subsMu.RLock()
			for _, s := range p.subscribers {
				if s.open {
					s.channel <- data
				}
			}
			p.subsMu.RUnlock()
		case "message":
			//Custom handling of SSE Message event
			fmt.Println("Encountered Message Event")
		default:
			// Unkown event encountered = discard
			continue
		}
	}
}

func identifyStreamData(message []byte) (event string, body []byte) {
	if message == nil {
		return "", nil
	}
	for i, b := range message {
		if b == ':' {
			event = string(message[:i])
			body = message[i+1:]
			break
		}
	}
	return
}

func decomposeDataEvent(text []byte) (data map[string]interface{}) {
	json.Unmarshal(text, &data)
	fmt.Println(data)
	return
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
		// Add this sub to the pub's list
		go sub.getSubData()
		// Start the clock!
		boom := time.After(dur)
		<-boom
		fmt.Printf("Time Channel Closed\n")
		pub.unSub(sub)

	} else {
		http.Error(w, "method not supported", 404)
	}
}

func (s *subscriber) getSubData() {
	for v := range s.channel {
		// sub will handle per-event processing here
		fmt.Println(v)
	}
}

func (p *publisher) addSubscriber(s *subscriber) {
	p.subsMu.Lock()
	p.subscribers = append(p.subscribers, s)
	p.subsMu.Unlock()
}

func (p *publisher) unSub(sub *subscriber) {
	//First 'lock' the sub so that it stops receiving
	sub.open = false
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
		make(chan map[string]any),
		true,
	}
	fmt.Printf("Created new sub with ID %v\n", s.id)
	return
}

// Create a new unique sub id using the timestamp.
// This timestamp is a mutex, ensuring uniqueness
//
// Could be an atomic counter, but oh well
func createNewSubId(mu *sync.Mutex) (id int64) {
	mu.Lock()
	id = time.Now().Unix()
	mu.Unlock()
	return
}

// func handleGetAnalysis(w http.ResponseWriter, r *http.Request) {

// 	if r.Method == "GET" {
// 		q, err := url.ParseQuery(r.URL.RawQuery)
// 		if err != nil {
// 			http.Error(w, "invalid query data", 400)
// 			return
// 		}
// 		dimension := q.Get("dimension")
// 		duration := q.Get("duration")
// 		if dimension == "" || duration == "" {
// 			http.Error(w, "invalid query value", 400)
// 			return
// 		}
// 		res, err := HandleAnalysisQuery(duration, dimension)
// 		if err != nil {
// 			http.Error(w, err.Error(), 500)
// 			return
// 		}
// 		jsonRes, err := json.Marshal(res)
// 		if err != nil {
// 			http.Error(w, "error creating response", 500)
// 			return
// 		}
// 		w.Header().Add("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(jsonRes)
// 	} else {
// 		http.Error(w, "method not supported", 404)
// 	}
// }
