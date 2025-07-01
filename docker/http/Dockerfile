FROM golang:1.24-alpine as dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD [ "go", "run", "./cmd/http" ]

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main-http ./cmd/http

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main-http .
CMD ["./main-http"]
