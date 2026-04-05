FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/npm-docker-auto-proxy ./cmd/npm-docker-auto-proxy

FROM alpine:3.21

COPY --from=builder /out/npm-docker-auto-proxy /usr/local/bin/npm-docker-auto-proxy

ENTRYPOINT ["/usr/local/bin/npm-docker-auto-proxy"]
