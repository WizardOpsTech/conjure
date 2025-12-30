#!/bin/bash
# Test script for conjure - provides easy access to testing commands

set -e

# Colors for output
GREEN='\033[0.32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default command
CMD=${1:-"all"}

case "$CMD" in
  "all")
    echo -e "${BLUE}Running all tests...${NC}"
    go test ./... -v
    ;;

  "unit")
    echo -e "${BLUE}Running unit tests only...${NC}"
    go test ./... -v -short
    ;;

  "integration")
    echo -e "${BLUE}Running integration tests...${NC}"
    go test ./cmd/bundle -v -run "TestGenerateBundle_"
    go test ./cmd/template -v -run "TestGenerateTemplate_"
    ;;

  "coverage")
    echo -e "${BLUE}Generating coverage report...${NC}"
    go test ./... -coverprofile=coverage.out -covermode=atomic
    go tool cover -func=coverage.out
    echo ""
    echo -e "${GREEN}Coverage report saved to coverage.out${NC}"
    echo "Run 'go tool cover -html=coverage.out' to view HTML report"
    ;;

  "coverage-html")
    echo -e "${BLUE}Generating HTML coverage report...${NC}"
    go test ./... -coverprofile=coverage.out -covermode=atomic
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}HTML coverage report saved to coverage.html${NC}"
    echo "Open coverage.html in your browser to view"
    ;;

  "watch")
    echo -e "${BLUE}Running tests in watch mode...${NC}"
    echo "Watching for changes... (Ctrl+C to stop)"
    while true; do
      go test ./... -count=1
      sleep 2
    done
    ;;

  "bench")
    echo -e "${BLUE}Running benchmarks...${NC}"
    go test ./... -bench=. -benchmem
    ;;

  "verbose")
    echo -e "${BLUE}Running tests with verbose output...${NC}"
    go test ./... -v -count=1
    ;;

  "help")
    echo "Usage: ./test.sh [command]"
    echo ""
    echo "Commands:"
    echo "  all              Run all tests (default)"
    echo "  unit             Run unit tests only"
    echo "  integration      Run integration tests only"
    echo "  coverage         Generate coverage report (terminal)"
    echo "  coverage-html    Generate HTML coverage report"
    echo "  watch            Run tests continuously on changes"
    echo "  bench            Run benchmarks"
    echo "  verbose          Run tests with verbose output"
    echo "  help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./test.sh                    # Run all tests"
    echo "  ./test.sh coverage           # Generate coverage report"
    echo "  ./test.sh coverage-html      # Generate HTML coverage report"
    ;;

  *)
    echo "Unknown command: $CMD"
    echo "Run './test.sh help' for usage information"
    exit 1
    ;;
esac
