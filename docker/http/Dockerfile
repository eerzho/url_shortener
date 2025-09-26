#BASE
FROM golang:1.25.1-alpine3.22 AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 80

#DEV
FROM base AS dev

RUN apk add --no-cache git

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN go install github.com/bufbuild/buf/cmd/buf@v1.57.2
RUN go install github.com/go-task/task/v3/cmd/task@v3.45.4
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0

COPY docker/http/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# CMD ["go", "run", "./cmd/http"]
ENTRYPOINT [ "/entrypoint.sh" ]

#BUILDER
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main-http ./cmd/http

#PROD
FROM gcr.io/distroless/static-debian12:nonroot AS prod

WORKDIR /app

COPY --from=builder /app/main-http .

USER nonroot:nonroot

CMD ["./main-http"]
