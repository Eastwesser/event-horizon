FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /
COPY pkg/migrator /pkg/migrator
COPY pkg/sqb /pkg/sqb
COPY platform /platform
COPY services/auth /services/auth
WORKDIR /services/auth

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o auth-service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /services/auth/auth-service /auth-service
EXPOSE 50051 9091
CMD ["/auth-service"]
