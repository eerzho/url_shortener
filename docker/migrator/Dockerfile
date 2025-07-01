FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o migrator ./cmd/migrator

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/migrator .
COPY migrations ./migrations
CMD ["./migrator", "-cmd=up"] 