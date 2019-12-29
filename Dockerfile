# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.13
ARG ASPROM_VERSION=1.8.0

# First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS build

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo '1000:x:65534:65534:1000:/:' > /user/passwd && \
    echo '1000:x:65534:' > /user/group

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
RUN apk add --no-cache ca-certificates git curl

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o asprom .

# Final stage: the running container.
FROM scratch AS final

# Import the user and group files from the build stage.
COPY --from=build /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /src/asprom .

EXPOSE 9145

CMD ["./asprom"]
