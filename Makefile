.PHONY: build install test test-unit test-coverage test-race test-bench test-clean test-watch

build:
	go build -o ./dist/docker-deliver ./cmd/docker-deliver

install:
	go install ./cmd/docker-deliver

test-unit:
	go test -v -race -timeout=5m ./internal/... ./cmd/...

test-coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./internal/... ./cmd/...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-race:
	go test -race -timeout=5m ./internal/... ./cmd/...

test-bench:
	go test -bench=. -benchmem ./internal/... ./cmd/...

test-clean:
	go clean -testcache

test-watch:
	./scripts/test-config.sh watch

test-all: test-clean test-unit test-coverage

e2e:
	go test ./test/e2e/...