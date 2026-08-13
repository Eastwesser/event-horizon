FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

# Preserve module replace paths: services/profile -> ../../pkg/migrator
WORKDIR /
COPY pkg/migrator /pkg/migrator
COPY services/profile /services/profile
WORKDIR /services/profile

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o profile-service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /services/profile/profile-service /profile-service
EXPOSE 50060 9099
CMD ["/profile-service"]
