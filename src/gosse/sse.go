// Copyright 2019 Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package gosse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	pack "github.com/mlavergn/gopack/src/gopack"
	"github.com/mlavergn/rxgo/src/rx"
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
func NewSSEPayload(data []byte, source string) *SSEPayload {
	lastEventID := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	return &SSEPayload{
		Type:        "message",
		Data:        data,
		LastEventID: lastEventID,
		Source:      source,
	}
}

// String provides a string representation of the struct suitable for logging
func (id *SSEPayload) String() string {
	return fmt.Sprintf("type: %s\ndata: %s\norigin: %s\nlastEventId: %s\nsource: %s", id.Type, string(id.Data), id.Origin, id.LastEventID, id.Source)
}

// SSE provides a []byte representation of the struct in SSE format
func (id *SSEPayload) SSE() []byte {
	return []byte(fmt.Sprintf("name:%s\ndata:%s\norigin:%s\nid:%s\nsource:%s\n\n", id.Type, string(id.Data), id.Origin, id.LastEventID, id.Source))
}

// JSON provides a string JSON representation of the struct
func (id *SSEPayload) JSON() string {
	result, _ := json.Marshal(id)
	return string(result)
}

// Decode provides a map representation of the data field
func (id *SSEPayload) Decode() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal(id.Data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// ParseSSEPayload parses raw stream data into SSEPayload structs
func ParseSSEPayload(lines [][]byte) *SSEPayload {
	payload := &SSEPayload{}
	for _, line := range lines {
		strline := string(line)
		if strings.HasPrefix(strline, "type:") {
			payload.Type = strline[5 : len(line)-1]
		} else if strings.HasPrefix(strline, "data:") {
			payload.Data = []byte(line[5 : len(line)-1])
		} else if strings.HasPrefix(strline, "origin:") {
			payload.Origin = strline[7 : len(line)-1]
		} else if strings.HasPrefix(strline, "lastEventId:") {
			payload.LastEventID = strline[12:]
		} else if strings.HasPrefix(strline, "source:") {
			payload.Source = strline[7 : len(line)-1]
		}
	}
	return payload
}

// -----------------------------------------------------------------------------
// SSEService

// SSEService service
type SSEService struct {
	listener *http.Server
	obs      *rx.Observable
	pack     *pack.Pack
}

// NewSSEService convenience init a Support instance
func NewSSEService(port int) *SSEService {
	hostPort := ":" + strconv.Itoa(port)
	id := &SSEService{
		listener: &http.Server{Addr: hostPort},
		obs:      nil,
	}

	http.Handle("/", http.HandlerFunc(id.handlerStatic))
	http.Handle("/events", http.HandlerFunc(id.handlerEvents))

	return id
}

func (id *SSEService) handlerStatic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if id.pack == nil {
		pipe, err := os.Open("static" + r.URL.String())
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(w, pipe)
	} else {
		pipe, err := id.pack.Pipe("static" + r.URL.String())
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(w, pipe)
	}
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

	sub := rx.NewSubscription()
	id.obs.Subscribe <- sub

	for {
		select {
		case <-r.Context().Done():
			sub.Complete <- true
			return
		case event := <-sub.Next:
			payload := rx.ToByteArray(event, nil)
			w.Write(payload)
			flush.Flush()
			break
		}
	}
}

// Start starts the http listener
func (id *SSEService) Start() {
	http := rx.NewHTTPRequest(0)
	subject, _ := http.SSESubject("http://express-eventsource.herokuapp.com/events", nil)
	subject.Map(func(event interface{}) interface{} {
		data := rx.ToByteArrayArray(event, nil)
		source, _ := os.Hostname()
		payload := ParseSSEPayload(data)
		payload.Source = source
		return payload.SSE()
	})
	id.obs = subject

	id.pack = pack.NewPack()
	_, err := id.pack.Load()
	if err != nil {
		id.pack = nil
	}

	// blocking
	id.listener.ListenAndServe()
}

// Stop starts the http listener
func (id *SSEService) Stop() {
	id.listener.Close()

	// halt backgound jobs
	id.obs.Complete <- true
}
