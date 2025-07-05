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

migrate-up:
	docker compose run --rm migrate up
.PHONY: migrate-up

migrate-down:
	docker compose run --rm migrate down $(if $(N),$(N),)
.PHONY: migrate-down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	docker compose run --rm migrate create -ext sql -dir /migrations -seq $(name)
.PHONY: migrate-create
