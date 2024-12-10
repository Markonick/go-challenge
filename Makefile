.PHONY: lint test build run install-lint

fmt:
	gofmt -w .
	goimports -w .

# Install golangci-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
lint:
	golangci-lint run --fix

# Run tests
test:
	go test -v ./...

# Run tests with gotestsum
test-sum:
	gotestsum

# Run tests with coverage
coverage:
	go test -cover ./...

# Build the binary
build:
	go build -o bin/hookbro cmd/hookbro/main.go
	
# Run the service
run:
	go run cmd/hookbro/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# New command for development with Air
install-air:
	go install github.com/air-verse/air@latest

# Dev command that uses Air for hot reloading
dev: install-air
	air
