# Testing Guide for Conjure

This document describes the testing strategy and practices for the Conjure project.

## Test Organization

```
conjure/
├── cmd/
│   ├── bundle/
│   │   ├── bundle_test.go                # Unit tests 
│   │   └── bundle_integration_test.go    # Integration tests 
│   ├── template/
│   │   ├── template_test.go              # Unit tests 
│   │   └── template_integration_test.go  # Integration tests 
│   └── list/
│       └── list_test.go                  # Unit tests
└── internal/
    └── metadata/
        └── metadata_test.go              # Unit tests 
```

## Running Tests

### Quick Start

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific package
go test ./cmd/template -v

# Run specific test
go test ./cmd/template -v -run TestGenerateTemplate_FullWorkflow
```

### Using Make (Recommended)

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate coverage report
make coverage

# Generate HTML coverage report
make coverage-html

# View help
make help
```

### Using Test Script (Linux/Mac)

```bash
# Run all tests
./test.sh

# Run with coverage
./test.sh coverage

# Generate HTML coverage
./test.sh coverage-html

# View help
./test.sh help
```

## Test Types

### Unit Tests

Unit tests test individual functions in isolation.

**Files**: `*_test.go` (e.g., `bundle_test.go`, `template_test.go`)

**What they test**:
- Variable parsing (`parseVariables`)
- Template rendering (`renderTemplate`)
- YAML parsing (`parseValues`)
- Metadata merging (`MergeVariablesForTemplate`)

**Example**:
```go
func TestParseVariables_Simple(t *testing.T) {
    varsList := []string{"app_name=myapp", "namespace=production"}
    result, err := parseVariables(varsList)

    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if result["app_name"] != "myapp" {
        t.Errorf("Expected app_name='myapp', got: %v", result["app_name"])
    }
}
```

### Integration Tests

Integration tests test complete workflows from start to finish.

**Files**: `*_integration_test.go`

**What they test**:
- Full template generation workflow (template file → variables → output file)
- Full bundle generation workflow (bundle directory → multiple templates → output directory)
- File I/O operations
- Config loading
- Error handling in real scenarios

**Example**:
```go
func TestGenerateTemplate_FullWorkflow(t *testing.T) {
    // Setup: Create temp environment with config
    baseDir, cleanup := setupTestEnvironment(t)
    defer cleanup()

    // Create template and values files
    // ... (file creation)

    // Execute the full workflow
    err := generateTemplate("deployment.yaml", outputPath, varsList, false, valuesPath)

    // Verify output file was created with correct content
    // ... (assertions)
}
```

## Coverage Reporting

### Generate Coverage Report

```bash
make coverage
```

Output:
```
github.com/thesudoYT/conjure/cmd/bundle    coverage: 72.6% of statements
github.com/thesudoYT/conjure/cmd/template  coverage: 77.7% of statements
github.com/thesudoYT/conjure/internal/metadata coverage: 51.7% of statements
total: (statements) 40.5%
```

### View HTML Coverage Report

```bash
make coverage-html
# Open coverage.html in your browser
```

The HTML report shows:
- Line-by-line coverage
- Covered lines (green)
- Uncovered lines (red)
- Partially covered lines (yellow)

### Coverage Goals

- **Critical paths**: 80%+ coverage (template/bundle generation)
- **Helper functions**: 70%+ coverage (parsing, validation)
- **Overall project**: 60%+ coverage

## Writing New Tests

### Guidelines

1. **Test behavior, not implementation**
   - Test what the function does, not how it does it
   - This makes tests resilient to refactoring

2. **Use table-driven tests for multiple scenarios**
   ```go
   tests := []struct {
       name     string
       input    string
       expected string
       wantErr  bool
   }{
       {"valid input", "key=value", "value", false},
       {"invalid input", "noequals", "", true},
   }
   ```

3. **Test both happy and error paths**
   - Test that valid inputs work correctly
   - Test that invalid inputs produce appropriate errors

4. **Use descriptive test names**
   - Format: `Test<FunctionName>_<Scenario>`
   - Example: `TestParseVariables_InvalidFormat`

5. **Keep tests independent**
   - Each test should set up and clean up its own state
   - Use `t.TempDir()` for temporary files
   - Use `defer cleanup()` for resource cleanup

### Adding a New Unit Test

```go
func TestNewFunction_Scenario(t *testing.T) {
    // Setup
    input := "test input"

    // Execute
    result, err := NewFunction(input)

    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if result != "expected" {
        t.Errorf("Expected 'expected', got: %s", result)
    }
}
```

### Adding a New Integration Test

```go
func TestGenerateX_FullWorkflow(t *testing.T) {
    // Setup test environment
    baseDir, cleanup := setupTestEnvironment(t)
    defer cleanup()

    // Create necessary files
    // ... (template, config, values files)

    // Execute the function
    err := generateX(...)
    if err != nil {
        t.Fatalf("generateX() failed: %v", err)
    }

    // Verify output
    output, _ := os.ReadFile(outputPath)
    if !strings.Contains(string(output), "expected content") {
        t.Errorf("Output should contain expected content")
    }
}
```

## Testing Best Practices

### DO:
- ✅ Write tests for new features before or alongside implementation
- ✅ Test error cases and edge cases
- ✅ Use meaningful test names that describe the scenario
- ✅ Keep tests simple and focused
- ✅ Use `t.TempDir()` for temporary files
- ✅ Clean up resources with `defer`

### DON'T:
- ❌ Test private implementation details
- ❌ Write tests that depend on external services
- ❌ Leave commented-out test code
- ❌ Write tests that depend on execution order
- ❌ Ignore failing tests

## Continuous Testing

### Pre-commit Testing

Run tests before committing:
```bash
make test
```

### Watch Mode (During Development)

```bash
./test.sh watch  # Runs tests continuously
```

### CI/CD Integration (Future)

Tests will automatically run on:
- Pull requests
- Commits to main branch
- Release tags

## What's Not Tested Yet

The following areas need test coverage:

### Highest Priority
- [ ] Interactive prompts (`internal/prompt/`)
  - Variable collection
  - Bubbletea UI interactions
- [ ] Config loading (`internal/config/`)

### Low Priority
- [ ] Cobra command initialization
- [ ] CLI flag handling

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Coverage Guide](https://go.dev/blog/cover)

## Questions?

If you have questions about testing:
1. Check this document
2. Look at existing tests for examples
3. Run `make help` for available commands
