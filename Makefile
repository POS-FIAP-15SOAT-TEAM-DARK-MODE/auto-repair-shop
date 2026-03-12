include .env
export

run:
	go run cmd/service/main.go

test:
	go test ./... --race -v

coverage:
	go test ./... --coverprofile=coverage.out
	go tool cover -html=coverage.out

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" down

migrate-status:
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" version
