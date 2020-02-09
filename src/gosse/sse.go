// Copyright 2019 Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package gosse

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// Version export
const Version = "0.3.2"

// -----------------------------------------------------------------------------
// SSEPayload

// SSEPayload type
type SSEPayload struct {
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	Origin      string                 `json:"origin"`
	LastEventID string                 `json:"lastEventId"`
	Source      string                 `json:"source"`
}

// NewSSEPayload init
func NewSSEPayload(data map[string]interface{}, origin string) *SSEPayload {
	lastEventID := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	hostname, _ := os.Hostname()
	return &SSEPayload{
		Type:        "message",
		Data:        data,
		Origin:      origin,
		LastEventID: lastEventID,
		Source:      hostname,
	}
}

// NewSSEPayloadFromRaw init
func NewSSEPayloadFromRaw(lines [][]byte) *SSEPayload {
	id := &SSEPayload{}
	for _, line := range lines {
		end := len(line) - 1
		strline := string(line)
		if strings.HasPrefix(strline, "type:") {
			id.Type = strline[5:end]
		} else if strings.HasPrefix(strline, "data:") {
			var data map[string]interface{}
			rawdata := []byte(line[5 : len(line)-1])
			json.Unmarshal(rawdata, &data)
			id.Data = data
		} else if strings.HasPrefix(strline, "origin:") {
			id.Origin = strline[7:end]
		} else if strings.HasPrefix(strline, "lastEventId:") {
			id.LastEventID = strline[12:end]
		} else if strings.HasPrefix(strline, "source:") {
			id.Source = strline[7:end]
		}
	}
	return id
}

// NewSSEPayloadFromRaw init
func NewSSEPayloadFromMap(dict map[string]interface{}) *SSEPayload {
	id := &SSEPayload{}
	for k, v := range dict {
		switch strings.ToLower(k) {
		case "type":
			id.Type = v.(string)
		case "data":
			id.Data = v.(map[string]interface{})
		case "origin":
			id.Origin = v.(string)
		case "lasteventid":
			id.LastEventID = v.(string)
		case "source":
			id.Source = v.(string)
		}
	}

	return id
}

// String provides a string representation of the struct suitable for logging
func (id *SSEPayload) String() string {
	var payload strings.Builder
	payload.WriteString("type:")
	payload.WriteString(id.Type)
	payload.WriteString("\ndata:")
	data, _ := json.Marshal(id.Data)
	payload.Write(data)
	payload.WriteString("\norigin:")
	payload.WriteString(id.Origin)
	payload.WriteString("\nlastEventId:")
	payload.WriteString(id.LastEventID)
	payload.WriteString("\nsource:")
	payload.WriteString(id.Source)

	return payload.String()
}

// SSE provides a []byte representation of the struct in SSE format
func (id *SSEPayload) SSE() []byte {
	var payload strings.Builder
	payload.WriteString("name:")
	payload.WriteString(id.Type)
	payload.WriteString("\ndata:")
	data, _ := json.Marshal(id.Data)
	payload.Write(data)
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
