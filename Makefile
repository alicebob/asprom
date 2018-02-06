.PHONY: all build test release

all: build test

build:
	go build -i

test:
	go test

release:
	goreleaser --rm-dist
