package gosse

import (
	"encoding/json"
	"math/rand"
	"time"
)

func dataSource() <-chan SSEPayloadData {
	ssePayloadDataChan := make(chan SSEPayloadData)

	// spin the SSEData feed
	go func() {
		data := map[string]map[string]string{
			"set1": {
				"val1": "hello",
				"val2": "world",
			},
			"set2": {
				"val1": "foo",
				"val2": "bar",
			},
		}
		keys := [3]string{}
		i := 0
		for k := range data {
			keys[i] = k
			i++
		}
		for {
			key := keys[rand.Intn(len(data))]
			payload, _ := json.Marshal(map[string]map[string]string{key: data[key]})
			ssePayloadDataChan <- payload
			time.Sleep(1 * time.Second)
		}
	}()

	return ssePayloadDataChan
}
