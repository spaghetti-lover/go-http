run:
	go run ./cmd/server
test:
	go test ./... -cover
check: test