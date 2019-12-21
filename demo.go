package main

import (
	"gosse"
)

func main() {
	service := gosse.NewSSEService(8000)
	service.Start()
}
