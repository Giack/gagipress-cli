.PHONY: build test test-integration test-coverage vet fmt clean install help

# Default target
all: build

# Build the binary
build:
	@echo "🔨 Building gagipress..."
	@mise exec -- go build -o bin/gagipress

# Run tests
test:
	@echo "🧪 Running tests..."
	@mise exec -- go test ./...

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	@mise exec -- go test ./test/integration/... -v

# Run test coverage
test-coverage:
	@echo "📊 Running test coverage..."
	@./scripts/test-coverage.sh

# Lint code
vet:
	@echo "🔍 Running go vet..."
	@mise exec -- go vet ./...

# Format code
fmt:
	@echo "✨ Formatting code..."
	@mise exec -- go fmt ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ dist/
	@rm -f coverage.out coverage.html

# Install globally
install: build
	@echo "📦 Installing to /usr/local/bin..."
	@sudo mv bin/gagipress /usr/local/bin/
	@echo "✅ Installed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build           - Build the binary (output: bin/gagipress)"
	@echo "  make test            - Run all tests"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make vet             - Run go vet"
	@echo "  make fmt             - Format code with go fmt"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make install         - Build and install globally"
	@echo "  make help            - Show this help message"
