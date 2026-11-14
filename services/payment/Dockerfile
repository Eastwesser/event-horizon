FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/payment .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o payment-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/payment-service /payment-service

EXPOSE 50058 9103

CMD ["/payment-service"]