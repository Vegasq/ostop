.PHONY: test coverage coverage-html lint build run clean

# Run all tests with race detection
test:
	go test -v -race ./...

# Run tests with coverage report
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

# Generate HTML coverage report
coverage-html:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Build the application
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ostop .

# Run the application
run:
	go run .

# Clean build artifacts and coverage files
clean:
	rm -f ostop coverage.out coverage.html
	go clean
