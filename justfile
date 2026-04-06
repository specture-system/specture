# Specture development justfile
# Run `just` to see available commands

export CGO_ENABLED := "0"

# Default: show help
default:
  @just --list

# Install Go dependencies
deps:
  go mod download

# Set up development environment
setup: deps
  pre-commit install

# Build the CLI binary
build:
  go build -o specture .

# Run the CLI with arguments (usage: just run setup --help)
[positional-arguments]
run *args:
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

check: format lint test

# Install the CLI locally
install:
  go install .

# Clean build artifacts
clean:
  rm -f specture
  go clean
