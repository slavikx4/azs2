MIGRATIONS_DIR = migrations
NAME_CONTAINER_DB = postgres-azs
DB_USER = postgres
DB_PASS = postgres
DB_HOST = localhost
DB_PORT = 5432
DB_NAME = azs
DB_URL = postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

run:
	go run ./cmd/main.go

build:
	go build -o bin/azs ./cmd/main.go

up_pg:
	docker run -d \
		--name $(NAME_CONTAINER_DB) \
		-e POSTGRES_PASSWORD=$(DB_PASS) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		postgres:17

start_pg:
	docker start $(NAME_CONTAINER_DB)

stop_pg:
	docker stop $(NAME_CONTAINER_DB)

delete_pg:
	docker rm -f $(NAME_CONTAINER_DB)

migrate_up:
	docker run --rm -v $(PWD)/$(MIGRATIONS_DIR):/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" up

migrate_down:
	docker run --rm -v $(PWD)/$(MIGRATIONS_DIR):/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" down

db_reset: delete_pg up_pg
	sleep 3
	docker run --rm -v $(PWD)/$(MIGRATIONS_DIR):/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)" up

.PHONY: run build up_pg start_pg stop_pg delete_pg migrate_up migrate_down db_reset