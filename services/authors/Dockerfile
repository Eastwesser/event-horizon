FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY services/authors .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o authors-service ./cmd/main.go

FROM scratch

COPY --from=builder /app/authors-service /authors-service

EXPOSE 50061 9104

CMD ["/authors-service"]