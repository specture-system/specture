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
  go build -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short=7 HEAD 2>/dev/null || echo unknown)" -o specture .

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
  go install -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short=7 HEAD 2>/dev/null || echo unknown)" .

# Clean build artifacts
clean:
  rm -f specture
  go clean
