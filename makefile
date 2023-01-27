.PHONY: test install build

OUT=bin/sync
IN=cmd/sync/sync.go

install:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod download

test:
	go test ./...

build:
	go build -o OUT ${IN}
