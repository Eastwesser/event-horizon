FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/game .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o game-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/game-service /game-service

EXPOSE 50052

CMD ["/game-service"]
