FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/analytics .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o analytics-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/analytics-service /analytics-service

EXPOSE 50057 9106

CMD ["/analytics-service"]