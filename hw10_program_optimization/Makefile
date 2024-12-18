format:
	gofumpt -w .
	golangci-lint run .

test: format
	go clean -testcache
	go test ./...
	go test -v -count=1 -timeout=30s -tags bench .

