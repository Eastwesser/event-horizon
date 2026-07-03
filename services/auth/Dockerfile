FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/auth .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o auth-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/auth-service /auth-service

EXPOSE 50051

CMD ["/auth-service"]
