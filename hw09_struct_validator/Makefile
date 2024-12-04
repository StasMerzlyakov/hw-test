format:
	gofumpt -w .
	golangci-lint run .

test: format
	go clean -testcache
	go test ./...
	go test -v -race -count=100 ./...

