# Specture development justfile
# Run `just` to see available commands

export CGO_ENABLED := "0"

# Default: show help
default:
  @just --list

# Build the CLI binary
build:
  go build -o specture .

# Run the CLI with arguments (usage: just run-dev setup --help)
[positional-arguments]
run-dev *args:
   go run . "$@"

# Run tests
test:
  go test -v ./...

# Format code
format:
  go fmt ./...

# Lint code with go vet
lint:
  go vet ./...

# Tidy dependencies
tidy:
  go mod tidy

# Install Go dependencies
deps:
  go mod download

# Set up development environment
setup: deps
  pre-commit install

# Check code (format, lint, test)
check: format lint test

# Install the CLI locally
install:
  go install .

# Clean build artifacts
clean:
  rm -f specture
  go clean
