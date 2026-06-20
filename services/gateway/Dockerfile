FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o gateway ./services/gateway/cmd/main.go

FROM scratch

COPY --from=builder /app/gateway /gateway

EXPOSE 8080

CMD ["/gateway"]
