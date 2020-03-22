package gosse

import (
	"fmt"
	"strings"
	"testing"
)

type SSEPayloadBench struct {
	Type        string `json:"type"`
	Data        string `json:"data"`
	Origin      string `json:"origin"`
	LastEventID string `json:"lastEventId"`
	Source      string `json:"source"`
}

var id = SSEPayloadBench{
	Type:        "message",
	Data:        "the quick brown dog jumped over the lazy fox",
	Origin:      "localhost",
	LastEventID: "123456",
	Source:      "localhost",
}

func BenchmarkFmtString(b *testing.B) {
	_ = fmt.Sprintf("type: %s\ndata: %s\norigin: %s\nid: %s\nsource: %s\n\n",
		id.Type, string(id.Data), id.Origin, id.LastEventID, id.Source)
}

// SSE provides a []byte representation of the struct in SSE format
func BenchmarkArrayJoin(b *testing.B) {
	data := []string{
		"name:", id.Type,
		"\ndata:", string(id.Data),
		"\norigin: ", id.Origin,
		"\nid: ", id.LastEventID,
		"\nsource: ", id.Source,
		"\n\n",
	}
	_ = strings.Join(data, "")
}

func BenchmarkBuilderString(b *testing.B) {
	var payload strings.Builder
	payload.WriteString("name:")
	payload.WriteString(id.Type)
	payload.WriteString("\ndata:")
	payload.WriteString(id.Data)
	payload.WriteString("\norigin:")
	payload.WriteString(id.Origin)
	payload.WriteString("\nid:")
	payload.WriteString(id.LastEventID)
	payload.WriteString("\nsource:")
	payload.WriteString(id.Source)
	payload.WriteString("\n\n")

	_ = payload.String()
}
