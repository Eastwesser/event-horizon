FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/history .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o history-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/history-service /history-service

EXPOSE 50062 9105

CMD ["/history-service"]