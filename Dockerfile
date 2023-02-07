FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go build -o udnsserver

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/udnsserver ./



