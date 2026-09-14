// Что значит каждая функция и
// Как будет собираться докер образ?
// Какие команды для сборки и запуска контейнера?
FROM golang:1.17 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o myapp .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/myapp .

CMD ["./myapp"]
