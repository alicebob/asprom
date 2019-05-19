.PHONY: all build test vendor release

all: build test

build:
	go build -mod=vendor

test:
	go test -mod=vendor

vendor:
	go mod vendor

release:
	goreleaser --rm-dist
