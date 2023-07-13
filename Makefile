.PHONY: dokcer.run
docker.run:
	docker run --name postgres-yandex-lavka -p 5432:5432  -e POSTGRES_USER=postgres  -e POSTGRES_PASSWORD=password -d postgres:alpine

.PHONY: dokcer.createdb
docker.createdb:
	docker exec -it postgres-yandex-lavka createdb --username=postgres --owner=postgres yandex_lavka_db

.PHONY: dokcer.dropdb
docker.dropdb:
	docker exec -it postgres-yandex-lavka dropdb --username=postgres yandex_lavka_db

.PHONY: migrate.install
migrate.install:
	brew install golang-migrate

.PHONY: migrate.create.tables
migrate.create.tables:
	migrate create -ext sql -dir internal/store/migrations -seq create_yandex_lavka_db_tables; \
	cat > internal/store/migrations/000001_create_yandex_lavka_db_tables.up.sql internal/store/db_data_model/data_model_table_create.sql; \
    cat > internal/store/migrations/000001_create_yandex_lavka_db_tables.down.sql internal/store/db_data_model/data_model_table_drop.sql

.PHONY: migrate.create.data
migrate.create.data:
	migrate create -ext sql -dir internal/store/migrations -seq create_yandex_lavka_db_data

.PHONY: migrate.up
migrate.up:
	migrate -path internal/store/migrations -database "postgresql://postgres:password@localhost:5432/yandex_lavka_db?sslmode=disable" -verbose up

.PHONY: migrate.down
migrate.down:
	migrate -path internal/store/migrations -database "postgresql://postgres:password@localhost:5432/yandex_lavka_db?sslmode=disable" -verbose down

.PHONY: migrate.up.default
migrate.up.default:
	migrate -path internal/store/migrations -database "postgresql://postgres:password@localhost:5432/postgres?sslmode=disable" -verbose up

.PHONY: migrate.down.default
migrate.down.default:
	migrate -path internal/store/migrations -database "postgresql://postgres:password@localhost:5432/postgres?sslmode=disable" -verbose down


.PHONY: mock.gen
mock.gen:
	mockgen -source=internal/domain/service/courier.go -destination=internal/store/postgressql/adapters/mocks/mock_courier.go
	mockgen -source=internal/domain/service/order.go -destination=internal/store/postgressql/adapters/mocks/mock_order.go


.PHONY: lint
lint:
	golangci-lint run


.PHONY: build
build:
	docker-compose build


.PHONY: run.docker.compose
run.docker.compose: build
	docker-compose up

.PHONY: stop.docker.compose
stop.docker.compose:
	docker-compose down


.PHONY: run.docker
run.docker:
	docker build -t yandex-lavka-service . && docker run -p 8080:8080 yandex-lavka-service
