FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/leaderboard .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o leaderboard-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/leaderboard-service /leaderboard-service

EXPOSE 50054

CMD ["/leaderboard-service"]
