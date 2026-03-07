run:
	go run cmd/service/main.go

test:
	go test ./... --race -v

coverage:
	go test ./... --coverprofile=coverage.out
	go tool cover -html=coverage.out
