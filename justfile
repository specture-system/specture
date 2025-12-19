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
run-dev *args:
  go run . {{args}}

# Run tests
test:
  go test -v ./...

# Format code
fmt:
  go fmt ./...

# Lint code with go vet
lint:
  go vet ./...

# Tidy dependencies
tidy:
  go mod tidy

# Check code (format, lint, test)
check: fmt lint test

# Install the CLI locally
install:
  go install .

# Clean build artifacts
clean:
  rm -f specture
  go clean
