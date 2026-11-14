FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/notification .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o notification-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/notification-service /notification-service

EXPOSE 50063 9102

CMD ["/notification-service"]