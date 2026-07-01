
.PHONY: fmt vet lint build test check

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test:
	go test ./...

build: vet
	go build ./...

lint:
	golangci-lint run

check: fmt vet lint test
