FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/services/balancer

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o balancer ./cmd/main.go

FROM scratch

COPY --from=builder /app/services/balancer/balancer /balancer

EXPOSE 8079

CMD ["/balancer"]