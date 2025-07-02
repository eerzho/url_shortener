FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
CMD ["go", "run", "./cmd/http"]

FROM base AS builder
RUN go build -o main-http ./cmd/http

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main-http .
CMD ["./main-http"]
