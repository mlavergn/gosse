package main

import (
	"gosse/src/gosse"
)

func main() {
	service := gosse.NewSSEService(8000)
	service.Start()
}
