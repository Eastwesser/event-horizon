FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /
COPY platform /platform
COPY services/nats-hub /services/nats-hub
WORKDIR /services/nats-hub

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /services/nats-hub/nats-hub /nats-hub
EXPOSE 9097
CMD ["/nats-hub"]
