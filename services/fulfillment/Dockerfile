FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/fulfillment .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o fulfillment-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/fulfillment-service /fulfillment-service

EXPOSE 9101

CMD ["/fulfillment-service"]