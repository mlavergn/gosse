###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

VERSION := 0.3.2

ver:
	@sed -i '' 's/^const Version = "[0-9]\{1,3\}.[0-9]\{1,3\}.[0-9]\{1,3\}"/const Version = "${VERSION}"/' sse.go

lint:
	$(shell go env GOPATH)/bin/golint .

build:
	go build .

demo: build
	cd cmd; go build -o ../demo demo.go service.go
	open "http://127.0.0.1:8000/index.html"
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
	go test -count=1 -v .

bench: build
	go test -benchmem -benchtime 10000x -bench=. -v .

release:
	zip -r gosse.zip LICENSE README.md Makefile cmd *.go mod.go
	hub release create -m "${VERSION} - gosse" -a gosse.zip -t master "v${VERSION}"
	open "https://github.com/mlavergn/gosse"

github:
	open "https://github.com/mlavergn/gosse"