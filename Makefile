
.PHONY: fmt vet lint build test check

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test:
	go test ./...

build: vet
	go build -o obsidian-mcp cmd/obsidian-mcp/main.go

lint:
	golangci-lint run

check: fmt vet lint test
