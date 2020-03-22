package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	pack "github.com/mlavergn/gopack"
	sse "github.com/mlavergn/gosse"
	rx "github.com/mlavergn/rxgo/src/rx"
)

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
	subject := rx.NewObservable().Merge(id.obs).Delay(1 * time.Second)
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
		source, _ := os.Hostname()
		payload := sse.NewSSEPayload(data, source)
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
