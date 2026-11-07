.PHONY: build run test lint clean install

APP=moreops
BINARY=bin/$(APP)

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/moreops

run:
	go run ./cmd/moreops

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/moreops

clean:
	rm -rf bin
