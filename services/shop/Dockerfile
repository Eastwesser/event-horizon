FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/shop .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o shop-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/shop-service /shop-service

EXPOSE 50059

CMD ["/shop-service"]
