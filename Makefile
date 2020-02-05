###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

VERSION := 0.1.0

ver:
	@sed -i '' 's/^const Version = "[0-9]\{1,3\}.[0-9]\{1,3\}.[0-9]\{1,3\}"/const Version = "${VERSION}"/' src/gosse/sse.go

lint:
	golint src/gosse

build:
	go build ./src/gosse/...

demo: build
	go build -o demo cmd/demo.go
	./demo

pack: demo
	zip pack static/*
	printf "%010d" `stat -f%z pack.zip` >> pack.zip
	mv demo main.pack; cat main.pack pack.zip > demo
	chmod +x demo
	rm pack.zip main.pack

www:
	open "http://127.0.0.1:8000/index.html"

events:
	open "http://127.0.0.1:8000/events"

test: build
	go test -v ./src/gosse/...

github:
	open "https://github.com/mlavergn/gosse"