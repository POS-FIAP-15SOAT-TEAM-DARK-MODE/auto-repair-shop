-include .env
export

OS := $(shell uname -s)

ifeq ($(OS),Windows_NT)
    SET_ENV_PREFIX := set
    ENV_SEPARATOR := &
    SHELL := cmd
    .SHELLFLAGS := /C
else
    SET_ENV_PREFIX := export
    ENV_SEPARATOR := ;
endif

run:
	go run cmd/service/main.go

test:
	$(SET_ENV_PREFIX) LOG_LEVEL=PANIC $(ENV_SEPARATOR) go test ./... --race -v

test-integration:
	$(SET_ENV_PREFIX) BCRYPT_COST=4 $(ENV_SEPARATOR) go test -tags=integration ./internal/integration/... -v

coverage:
	$(SET_ENV_PREFIX) LOG_LEVEL=PANIC $(ENV_SEPARATOR) go test ./... --coverprofile=coverage.out
	go tool cover -html=coverage.out

sonarqube-run: coverage
	sonar-scanner \
	  -Dsonar.projectKey=auto-repair-shop \
	  -Dsonar.sources=. \
	  -Dsonar.host.url=http://localhost:9000 \
	  -Dsonar.token=${SONARQUBE_PROJECT_TOKEN}

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" up

migrate-down:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" down

migrate-status:
	@type migrate >/dev/null 2>&1 || { $(MAKE) migrate-install; }
	migrate -path migrations -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/autorepairshop?sslmode=disable" version

mockgen:
	go generate ./...
