FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
CMD ["go", "run", "./cmd/migrator", "-cmd=up"]

FROM base AS builder
RUN go build -o migrator ./cmd/migrator

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/migrator .
COPY migrations ./migrations
CMD ["./migrator", "-cmd=up"]
