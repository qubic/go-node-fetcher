FROM golang:1.21 AS builder
ENV CGO_ENABLED=0

WORKDIR /src
COPY . /src

RUN go build -o "/src/bin/go-node-fetcher"

# We don't need golang to run binaries, just use alpine.
FROM alpine:latest
COPY --from=builder /src/bin/go-node-fetcher /app/go-node-fetcher
RUN chmod +x /app/go-node-fetcher

EXPOSE 8080

WORKDIR /app

ENTRYPOINT ["./go-node-fetcher"]