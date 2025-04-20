include .env

build:
	@go build -o bin/start cmd/app/main.go

run: build
	@./bin/start

#test-integrat:
#	@go test ./test/integration/... -v
#
#test-integrat-cov:
#	@go test ./test/integration/... -cover -v

migrate-create:
	migrate create -ext sql -dir database/migrations $(name)

migrate-up:
	migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" -path database/migrations up

migrate-down:
	migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" -path database/migrations down

