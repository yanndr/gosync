VERSION=$(shell cat .version)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
BINARY=sync

.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY} ./cmd/sync/main.go

.PHONY: test
test:
	go test -race ./...