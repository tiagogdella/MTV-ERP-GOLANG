.PHONY: db migrate catalog gateway test stop

DB_URL := postgres://catalog:catalog@localhost:5432/catalog?sslmode=disable

db:
	docker start catalog-db-local

migrate:
	migrate -path catalog-service/migrations -database "$(DB_URL)" up

catalog:
	cd catalog-service && DATABASE_URL="$(DB_URL)" PORT=8081 GRPC_PORT=9090 go run ./cmd/server

gateway:
	cd api-gateway && CATALOG_SERVICE_ADDR=localhost:9090 go run ./cmd/server

test:
	cd api-gateway && go test ./internal/handlers/ -v

stop:
	docker stop catalog-db-local
