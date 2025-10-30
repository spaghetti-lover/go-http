run:
	go run ./cmd/server
lint:
	golangci-lint run
test:
	go test ./... -cover