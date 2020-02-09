package main

func main() {
	service := NewSSEService(8000)
	service.Start()
}
