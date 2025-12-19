# Specture development justfile
# Run `just` to see available commands

export CGO_ENABLED := "0"

# Default: show help
default:
  @just --list

# Build the CLI binary
build:
  go build -o specture .

# Run tests
test:
  go test -v ./...

# Run tests with coverage
coverage:
  go test -v -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out

# Format code
fmt:
  go fmt ./...

# Lint code with go vet
lint:
  go vet ./...

# Tidy dependencies
tidy:
  go mod tidy

# Run the CLI with arguments (usage: just run setup --help)
run *args:
  go run . {{args}}

# Install the CLI locally
install:
  go install .

# Clean build artifacts
clean:
  rm -f specture
  go clean
