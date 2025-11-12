run:
	go run ./cmd/server
lint:
	golangci-lint run --config=.golangci.yml
test:
	go test ./... -cover
check: lint test