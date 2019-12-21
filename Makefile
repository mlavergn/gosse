###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

GOPATH = "${PWD}"

lint:
	GOPATH=${GOPATH} ~/go/bin/golint sse.go

build:
	GOPATH=${GOPATH} go build .

demo: build
	GOPATH=${GOPATH} go build -o demo demo.go
	./demo

www:
	open "http://127.0.0.1:8000/events"

test: build
	GOPATH=${GOPATH} go test -v .