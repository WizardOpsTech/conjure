.PHONY: test test-unit test-integration test-verbose coverage coverage-html clean help

GO := $(shell command -v go)
ifeq ($(GO),)
$(error go not found in PATH)
endif

# Default target
all: test

# Run all tests
test:
	@echo "Running all tests..."
	@$(GO) test ./... -count=1

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@$(GO) test ./cmd/bundle -v -run "TestParse|TestRender"
	@$(GO) test ./cmd/template -v -run "TestParse|TestRender"
	@$(GO) test ./internal/metadata -v

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@$(GO) test ./cmd/bundle -v -run "TestGenerateBundle_"
	@$(GO) test ./cmd/template -v -run "TestGenerateTemplate_"

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@$(GO) test ./... -v -count=1

# Generate coverage report (terminal)
coverage:
	@echo "Generating coverage report..."
	@$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	@echo ""
	@echo "Coverage by package:"
	@$(GO) tool cover -func=coverage.out | grep total:
	@echo ""
	@echo "Detailed coverage saved to coverage.out"
	@echo "Run 'make coverage-html' to generate HTML report"

# Generate HTML coverage report
coverage-html:
	@echo "Generating HTML coverage report..."
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "HTML coverage report saved to coverage.html"
	@echo "Open coverage.html in your browser to view"

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@$(GO) clean -testcache

# Show help
help:
	@echo "Conjure Test Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  test             Run all tests (default)"
	@echo "  test-unit        Run unit tests only"
	@echo "  test-integration Run integration tests only"
	@echo "  test-verbose     Run tests with verbose output"
	@echo "  coverage         Generate coverage report (terminal)"
	@echo "  coverage-html    Generate HTML coverage report"
	@echo "  clean            Remove test artifacts and cache"
	@echo "  help             Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Run all tests"
	@echo "  make coverage           # Generate coverage report"
	@echo "  make coverage-html      # Generate HTML coverage report"
