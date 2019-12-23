// Copyright 2019 Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package gosse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------
// SSEPayload

// SSEPayloadData payload
type SSEPayloadData []byte

// SSEPayload type
type SSEPayload struct {
	Type        string         `json:"type"`
	Data        SSEPayloadData `json:"data"`
	Origin      string         `json:"origin"`
	LastEventID string         `json:"lastEventId"`
	Source      string         `json:"source"`
}

// NewSSEPayload conventience init
func NewSSEPayload(data []byte) SSEPayload {
	lastEventID := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	return SSEPayload{
		Type:        "message",
		Data:        data,
		LastEventID: lastEventID,
	}
}

// String provides a string representation of the struct suitable for logging
func (id SSEPayload) String() string {
	return fmt.Sprintf("type: %s\ndata: %s\norigin: %s\nlastEventId: %s\nsource: %s", id.Type, string(id.Data), id.Origin, id.LastEventID, id.Source)
}

// SSE provides a []byte representation of the struct in SSE format
func (id SSEPayload) SSE() []byte {
	return []byte(fmt.Sprintf("name:%s\ndata:%s\norigin:%s\nid:%s\nsource:%s\n\n", id.Type, string(id.Data), id.Origin, id.LastEventID, id.Source))
}

// JSON provides a string JSON representation of the struct
func (id SSEPayload) JSON() string {
	result, _ := json.Marshal(id)
	return string(result)
}

// -----------------------------------------------------------------------------
// SSESubject

// ServiceSubject type
type ServiceSubject struct {
	Observable  chan SSEPayloadData
	Observers   map[chan SSEPayloadData]chan SSEPayloadData
	Close       chan bool
	Subscribe   chan ServiceSubject
	Unsubscribe chan ServiceSubject
}

// NewServiceSubject convenience init
func NewServiceSubject() ServiceSubject {
	id := ServiceSubject{
		Observable:  make(chan SSEPayloadData),
		Observers:   map[chan SSEPayloadData]chan SSEPayloadData{},
		Close:       make(chan bool, 1),
		Subscribe:   make(chan ServiceSubject, 1),
		Unsubscribe: make(chan ServiceSubject, 1),
	}

	// subjects multicast
	id.multicast()

	return id
}

// NewServiceObserver convenience init
func NewServiceObserver() ServiceSubject {
	return ServiceSubject{
		Observable: make(chan SSEPayloadData),
		Close:      make(chan bool, 1),
	}
}

// finalize helper closes the contained channels
func (id ServiceSubject) finalize() {
	close(id.Observable)
	close(id.Close)
}

// close helper closes the contained channels
func (id ServiceSubject) complete() {
	id.Close <- true
}

// multicast to all observers
func (id ServiceSubject) multicast() {
	go func() {
		for {
			select {
			case sub := <-id.Subscribe:
				id.Observers[sub.Observable] = sub.Observable
				break
			case sub := <-id.Unsubscribe:
				delete(id.Observers, sub.Observable)
				sub.finalize()
				break
			case event := <-id.Observable:
				for _, sub := range id.Observers {
					sub <- event
				}
				break
			}
		}
	}()
}

// -----------------------------------------------------------------------------
// SSEService

// SSEService service
type SSEService struct {
	listener *http.Server
	obs      ServiceSubject
}

// NewSSEService convenience init a Support instance
func NewSSEService(port int) *SSEService {
	hostPort := ":" + strconv.Itoa(port)
	id := &SSEService{
		listener: &http.Server{Addr: hostPort},
		obs:      NewServiceSubject(),
	}

	http.Handle("/", http.HandlerFunc(id.handlerStatic))
	http.Handle("/events", http.HandlerFunc(id.handlerEvents))

	return id
}

func (id *SSEService) handlerStatic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	data, err := ioutil.ReadFile("static" + r.URL.String())
	if err != nil {
		log.Println(err)
		return
	}

	w.Write(data)
}

func (id *SSEService) handlerEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "GoSSE Server")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	lastEventID := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	w.Header().Set("Last-Event-ID", lastEventID)

	w.WriteHeader(http.StatusOK)

	var flush http.Flusher
	if flusher, ok := w.(http.Flusher); ok {
		flush = flusher
	}

	sub := NewServiceObserver()
	id.obs.Subscribe <- sub

	for {
		select {
		case <-r.Context().Done():
			id.obs.Unsubscribe <- sub
			return
		case event := <-sub.Observable:
			w.Write(NewSSEPayload(event).SSE())
			flush.Flush()
			break
		}
	}
}

// Start starts the http listener
func (id *SSEService) Start() {
	// spin up the SSEPayloadData feed
	go func() {
		ssePayloadDataChan := dataSource()
		for {
			select {
			case payload := <-ssePayloadDataChan:
				id.obs.Observable <- payload
				break
			}
		}
	}()

	// blocking
	id.listener.ListenAndServe()
}

// Stop starts the http listener
func (id *SSEService) Stop() {
	id.listener.Close()

	// halt backgound jobs
	id.obs.complete()
	id.obs.finalize()
}
