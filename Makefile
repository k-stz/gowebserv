.DEFAULT_GOAL := run

.PHONY: fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/gowebserv

test:
	go test

run: build
	./bin/gowebserv