ifneq (,$(wildcard ./.env))
	include .env
	export
endif

.PHONY: up down db-clean run migrate-up migrate-down migrate-action migrate-create

up:
	docker compose up -d

down:
	docker compose down

db-clean:
	docker compose down -v
	@echo "Данные БД полностью очищены"

run:
	go run cmd/user-service/main.go

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action="down 1"

# Выполняем миграцию внутри докер-контейнера.
# Обрати внимание: имя сервиса БД должно совпадать с тем, что в docker-compose.yml (user-postgres или user-db)
migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "You need to identify 'action' variable. Example: make migrate-action action=up"; \
		exit 1; \
	fi ; \
	docker compose run --rm user-migrator \
		-path=//migrations \
		-database="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@user-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
		$(action)

# Команда для генерации новых файлов миграций через Docker (без локальной установки migrate)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Ошибка: укажите имя миграции. Пример: make migrate-create name=add_email"; \
		exit 1; \
	fi
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate:v4.19.1 create -ext sql -dir /migrations -seq $(name)