# Specture development justfile
# Run `just` to see available commands

# Default: show help
default:
  @just --list

# Build the CLI binary
build:
  CGO_ENABLED=0 go build -o specture .

# Run tests
test:
  CGO_ENABLED=0 go test -v ./...

# Run tests with coverage
coverage:
  CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out

# Format code
fmt:
  go fmt ./...

# Lint code with go vet
lint:
  CGO_ENABLED=0 go vet ./...

# Tidy dependencies
tidy:
  go mod tidy

# Run the CLI with arguments (usage: just run setup --help)
run *args:
  CGO_ENABLED=0 go run . {{args}}

# Install the CLI locally
install:
  CGO_ENABLED=0 go install .

# Clean build artifacts
clean:
  rm -f specture
  go clean
