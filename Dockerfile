# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth-service .
COPY ./swagger ./swagger

EXPOSE 8080

CMD ["./auth-service"]