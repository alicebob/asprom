.PHONY: all build test release

all: build test

build:
	go build

test:
	go test

release:
	goreleaser --rm-dist
