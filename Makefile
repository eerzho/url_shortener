build:
	docker compose build
.PHONY: build

up:
	docker compose up -d
.PHONY: up

down:
	docker compose down
.PHONY: down

swag:
	docker compose exec http swag init -g cmd/http/main.go
.PHONY: swag
