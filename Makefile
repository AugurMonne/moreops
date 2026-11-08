.PHONY: build run test lint clean install dist

APP=moreops
BINARY=bin/$(APP)

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/moreops

run:
	go run ./cmd/moreops

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/moreops

clean:
	rm -rf bin dist

dist:
	mkdir -p dist
	# Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/linux_amd64/moreops ./cmd/moreops
	tar -czf dist/moreops_linux_amd64.tar.gz -C bin/linux_amd64 moreops
	# Linux arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/linux_arm64/moreops ./cmd/moreops
	tar -czf dist/moreops_linux_arm64.tar.gz -C bin/linux_arm64 moreops
	# macOS arm64 (Apple Silicon)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/darwin_arm64/moreops ./cmd/moreops
	tar -czf dist/moreops_darwin_arm64.tar.gz -C bin/darwin_arm64 moreops
	# macOS amd64 (Intel)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/darwin_amd64/moreops ./cmd/moreops
	tar -czf dist/moreops_darwin_amd64.tar.gz -C bin/darwin_amd64 moreops
	# Windows amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/windows_amd64/moreops.exe ./cmd/moreops
	tar -czf dist/moreops_windows_amd64.tar.gz -C bin/windows_amd64 moreops.exe
	@echo "✓ Release archives generated in dist/"

