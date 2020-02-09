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
func NewSSEPayload(data []byte, origin string) *SSEPayload {
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

// String provides a string representation of the struct suitable for logging
func (id *SSEPayload) String() string {
	var payload strings.Builder
	payload.WriteString("type:")
	payload.WriteString(id.Type)
	payload.WriteString("\ndata:")
	payload.Write(id.Data)
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
