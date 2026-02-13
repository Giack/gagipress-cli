#!/bin/bash
set -e

echo "🧪 Running tests with coverage..."
echo ""

# Run unit tests with coverage
mise exec -- go test ./internal/... ./cmd/... -coverprofile=coverage.out -covermode=atomic

echo ""
echo "📊 Generating coverage report..."

# Generate HTML report
mise exec -- go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
echo ""
echo "📈 Coverage Summary:"
echo "══════════════════════════════════════"
mise exec -- go tool cover -func=coverage.out | grep total

echo ""
echo "✅ Coverage report generated!"
echo "   • Text report: coverage.out"
echo "   • HTML report: coverage.html"
echo ""
echo "To view HTML report: open coverage.html"
