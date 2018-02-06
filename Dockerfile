# builder
FROM golang:1.9.3-alpine3.7 AS builder

WORKDIR /go/src/github.com/alicebob

ARG ASPROM_VERSION=1.0.1

RUN apk add --no-cache git \
    && git clone https://github.com/alicebob/asprom.git \
    && cd asprom \
    && git fetch --all --tags --prune \
    && git checkout tags/$ASPROM_VERSION

WORKDIR /go/src/github.com/alicebob/asprom

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o asprom .



# final
FROM alpine:3.7

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/github.com/alicebob/asprom .

EXPOSE 9145

CMD ["./asprom"]
