default: fmt lint test

build:
	go build -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

bench:
	go test -bench=. -benchtime=5s -run=^$ -benchmem

cgo-bench:
	go test -bench=CGO -benchtime=5s -run=^$ -benchmem

no-cgo-bench:
	CGO_ENABLED=0 go test -bench=CGO -benchtime=5s -run=^$ -benchmem

.PHONY: fmt lint test build bench cgo-bench no-cgo-bench
