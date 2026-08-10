FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/nats-hub .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub ./main.go

FROM scratch

COPY --from=builder /app/nats-hub /nats-hub

EXPOSE 9097

CMD ["/nats-hub"]