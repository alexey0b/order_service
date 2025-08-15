# Env for Goose migration:
MIGRATIONS_DIR?=internal/request/repositoriy/migrations

# Env for docker:
APP_CONTAINER=app
KAFKA_CONTAINER=broker
POSTGRES_CONTAINER=postgres

# Env for Postgres:
DB_USER?=user
DB_PASSWORD?=password
DB_HOST?=localhost
DB_PORT?=5432
DB_NAME?=my_db
DB_SSLMODE?=disable
DB_DSN="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

# Env for Kafka:
KAFKA_HOST=$(KAFKA_CONTAINER)
KAFKA_PORT?=9092

# Commands for Order service:
app-start:
	@docker compose up -d app

app-stop:
	@docker compose stop app

broker-start:
	@docker compose up -d broker

broker-stop:
	@docker compose down broker
	
postgres-start:
	@docker compose up -d postgres

postgres-stop:
	@docker compose stop postgres

service-start:
	@docker compose up -d

service-stop:
	@docker compose down

# Commands for Goose migration:
install-goose:
	@go install github.com/pressly/goose/v3/cmd/goose@latest

new-migration:
ifndef NAME
	$(error Usage: make new-migration NAME='migration_name')
endif
	@goose -dir $(MIGRATIONS_DIR) create -s $(NAME) sql


migrate-up:
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

migrate-down:
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) down

migrate-reset:
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) reset

migrate-status:
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) status

# Commands for Postgres:
postgres-create-user:
ifndef NAME
	$(error Usage: make postgres-create-user NAME=username PASSWORD=password)
endif
ifndef PASSWORD
	$(error Usage: make postgres-create-user NAME=username PASSWORD=password)
endif
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "CREATE USER $(NAME) WITH PASSWORD '$(PASSWORD)' LOGIN;"

postgres-grant-permissions:
ifndef NAME
	$(error Usage: make postgres-grant-permissions NAME=username)
endif
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "GRANT ALL PRIVILEGES ON DATABASE $(DB_NAME) TO $(NAME);"
	@docker exec -it $(POSTGRES_CONTAINER) psql -U  $(DB_USER) -d $(DB_NAME) -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $(NAME);"
	@docker exec -it $(POSTGRES_CONTAINER) psql -U  $(DB_USER) -d $(DB_NAME) -c "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $(NAME);"

# Commands for Kafka:
broker-create-topic:
ifndef NAME
	$(error Usage: make broker-create-topic NAME=topic_name)
endif
	@docker exec -it $(KAFKA_CONTAINER) ./opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT) --create --topic $(NAME)

broker-list-topics:
	@docker exec -it $(KAFKA_CONTAINER) ./opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT) --list

# Sends messages to topic:
broker-send-msgs:
	@docker exec -it $(APP_CONTAINER) ./producer

# Commands for tests:
unit-test-start:
	@echo "Запуск тестов для handlers:"
	@go test -v ./internal/delivery/rest/handler_test.go

	@echo "Запуск тестов для services:"
	@go test -v ./internal/usecase/service_test.go

	@echo "Запуск тестов для LRU:"
	@go test -v ./internal/infrastructure/cache/lru_test.go
	
integration-test-start:
	@echo "Запуск тестов для Postgres:"
	@go test -v ./internal/request/repositoriy/postgres/request_test.go

help:
	@echo "Available commands:"
	@echo ""
	@echo "For Order Service:"
	@echo "  service-start                - Start all services"
	@echo "  service-stop                 - Stop all services"
	@echo "  app-start                    - Start app container"
	@echo "  app-stop                     - Stop app container"
	@echo "  postgres-start               - Start postgres container"
	@echo "  postgres-stop                - Stop postgres container"
	@echo "  broker-start                 - Start broker container"
	@echo "  broker-stop                  - Stop broker container"
	@echo ""
	@echo "For Goose migration:"
	@echo "  install-goose                - Install goose migration tool"
	@echo "  new-migration NAME=...       - Create new migration file"
	@echo "  migrate-up                   - Apply all pending migrations"
	@echo "  migrate-down                 - Roll back the last migration"
	@echo "  migrate-reset                - Roll back ALL migrations (clean database)"
	@echo "  migrate-status               - Show migration status"
	@echo ""
	@echo "For Postgres:"
	@echo "  postgres-create-user NAME=... PASSWORD=... - Create user"
	@echo "  postgres-grant-permissions NAME=...        - Grant permissions to user"
	@echo ""
	@echo "For Kafka:"
	@echo "  broker-create-topic NAME=... - Create topic"
	@echo "  broker-list-topics           - Show all topics"
	@echo "  broker-send-msgs             - Send test messages"
	@echo ""
	@echo "For Tests:"
	@echo "  unit-test-start              - Run unit tests (handlers, services, cache)"
	@echo "  integration-test-start       - Run integration tests (postgres repository)"

.PHONY: help app-start app-stop postgres-start postgres-stop broker-start broker-stop service-start service-stop install-goose new-migration migrate-up migrate-down migrate-reset migrate-status postgres-create-user postgres-grant-permissions broker-create-topic broker-list-topics broker-send-msgs unit-test-start integration-test-start
