.PHONY: all build test

all: build test

build:
	go build -i

test:
	go test
