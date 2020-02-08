// Copyright 2019 Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package gosse

import (
	"encoding/json"
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

// Version export
const Version = "0.1.0"

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
	result := []string{
		"type: ", id.Type,
		"\ndata: ", string(id.Data),
		"\norigin: ", id.Origin,
		"\nlastEventId: ", id.LastEventID,
		"\nsource: ", id.Source,
	}
	return strings.Join(result, "")
}

// SSE provides a []byte representation of the struct in SSE format
func (id *SSEPayload) SSE() []byte {
	var payload strings.Builder
	payload.WriteString("name:")
	payload.WriteString(id.Type)
	payload.WriteString("\ndata:")
	payload.Write(id.Data)
	payload.WriteString("\norigin:")
	payload.WriteString(id.Origin)
	payload.WriteString("\nid:")
	payload.WriteString(id.LastEventID)
	payload.WriteString("\nsource:")
	payload.WriteString(id.Source)
	payload.WriteString("\n\n")

	return []byte(payload.String())
}

// JSON provides a string JSON representation of the struct
func (id *SSEPayload) JSON() string {
	result, _ := json.Marshal(id)
	return string(result)
}

// Decode provides a map representation of the data field
func (id *SSEPayload) Decode() (map[string]interface{}, error) {
	result := map[string]interface{}{}
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
		end := len(line) - 1
		strline := string(line)
		if strings.HasPrefix(strline, "type:") {
			payload.Type = strline[5:end]
		} else if strings.HasPrefix(strline, "data:") {
			payload.Data = []byte(line[5:end])
		} else if strings.HasPrefix(strline, "origin:") {
			payload.Origin = strline[7:end]
		} else if strings.HasPrefix(strline, "lastEventId:") {
			payload.LastEventID = strline[12:end]
		} else if strings.HasPrefix(strline, "source:") {
			payload.Source = strline[7:end]
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

func (id *SSEService) handlerStatic(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	resp.Header().Set("Content-Type", "text/html")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "close")
	resp.WriteHeader(http.StatusOK)

	if id.pack == nil {
		pipe, err := os.Open("static" + req.URL.String())
		if err != nil {
			log.Println(err)
			return
		}
		defer pipe.Close()
		io.Copy(resp, pipe)
	} else {
		pipe, err := id.pack.Pipe("static" + req.URL.String())
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(resp, pipe)
	}
}

func (id *SSEService) handlerEvents(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	resp.Header().Set("Server", "GoSSE Server")
	resp.Header().Set("Content-Type", "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "close")

	lastEventID := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	resp.Header().Set("Last-Event-ID", lastEventID)

	resp.WriteHeader(http.StatusOK)

	var flush http.Flusher
	if flusher, ok := resp.(http.Flusher); ok {
		flush = flusher
	}

	observer := rx.NewObserver()
	subject := rx.NewSubject().Merge(id.obs).Delay(1 * time.Second)
	subject.Subscribe <- observer

	for {
		select {
		case <-req.Context().Done():
			observer.Complete <- nil
			return
		case event := <-observer.Next:
			payload := rx.ToByteArray(event, nil)
			resp.Write(payload)
			flush.Flush()
			break
		}
	}
}

// Start starts the http listener
func (id *SSEService) Start() {
	data := map[string]interface{}{
		"set1": map[string]interface{}{
			"val1": "hello",
			"val2": "world",
		},
		"set2": map[string]interface{}{
			"val1": "foo",
			"val2": "bar",
		},
		"set3": map[string]interface{}{
			"val1": "tic",
			"val2": "toc",
		},
		"set4": map[string]interface{}{
			"val1": "zom",
			"val2": "bee",
		},
	}

	id.obs = rx.NewReplaySubject(4).Map(func(event interface{}) interface{} {
		if event == nil {
			return nil
		}
		data := event.(map[string]interface{})
		json, _ := json.Marshal(data)
		source, _ := os.Hostname()
		payload := NewSSEPayload(json, source)
		return payload.SSE()
	}).Share()
	id.obs.UID = "SSESubject" + id.obs.UID
	dataObs := rx.NewFromMap(data)
	dataObs.UID = "DataSubject" + dataObs.UID
	dataObs.Pipe(id.obs)

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
	id.obs.Complete <- id.obs
}
