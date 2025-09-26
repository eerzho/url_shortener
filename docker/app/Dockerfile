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

COPY docker/app/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

#BUILDER
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./cmd/app/main ./cmd/app/main.go

#PROD
FROM gcr.io/distroless/static-debian12:nonroot AS prod

WORKDIR /app

COPY --from=builder /app/cmd/app/main .

USER nonroot:nonroot

CMD ["./main"]
