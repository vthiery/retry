all: fmt lint test

.PHONY: fmt
fmt:
	go mod tidy
	gofumpt -s -w .
	gofumports -w .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	mkdir -p bin
	go test -race -coverprofile=bin/cover.out ./...
	go tool cover -html=bin/cover.out -o bin/cover.html
