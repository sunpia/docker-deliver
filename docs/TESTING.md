# Go Test Configuration

This document describes the comprehensive test configuration setup for the docker-deliver project.

## Overview

The project includes multiple test configuration files and tools to provide a robust testing environment:

- **Unit Tests**: Fast, isolated tests for individual functions and packages
- **Integration Tests**: Tests that verify component interactions
- **End-to-End Tests**: Full workflow tests
- **Coverage Reports**: Code coverage analysis and reporting
- **CI/CD Integration**: GitHub Actions workflow for automated testing

## Quick Start

### Running Tests

```bash
# Run all unit tests
make test-unit

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run benchmark tests
make test-bench

# Clean test cache
make test-clean
```

### Using the Test Script

```bash
# Run unit tests with verbose output
./scripts/test-config.sh unit -v

# Run coverage analysis
./scripts/test-config.sh coverage

# Run all tests
./scripts/test-config.sh all

# Clean test cache
./scripts/test-config.sh clean
```

## Configuration Files

### 1. `.testconfig`
Central test configuration file with default settings:
- Test timeout: 5 minutes
- Coverage mode: atomic
- Coverage threshold: 70%
- Race detection: enabled

### 2. VS Code Configuration (`.vscode/`)

#### `settings.json`
- Go test flags and timeout settings
- Coverage highlighting configuration
- Build and lint on save settings

#### `tasks.json`
Pre-configured VS Code tasks:
- **Go: Test All**: Run all tests
- **Go: Test Unit**: Run unit tests only
- **Go: Test with Coverage**: Generate coverage reports
- **Go: Generate Coverage Report**: Create HTML coverage report
- **Go: Test Current Package**: Test current package only
- **Go: Benchmark Tests**: Run performance benchmarks
- **Go: Clean Test Cache**: Clear Go test cache

#### `launch.json`
Debug configurations for:
- Debug current test function
- Debug package tests
- Debug all tests
- Debug specific package tests (compose, commands)

### 3. GitHub Actions (`.github/workflows/test.yml`)
Automated CI/CD pipeline that:
- Runs tests on multiple Go versions (1.21-1.24)
- Performs code linting and vetting
- Generates coverage reports
- Uploads coverage to Codecov
- Builds and tests installation

### 4. Test Script (`scripts/test-config.sh`)
Comprehensive bash script providing:
- Multiple test commands (unit, integration, e2e, all)
- Configurable options (verbose, timeout, race detection)
- Coverage analysis with threshold checking
- File watching for continuous testing
- Colored output for better readability

## Test Organization

### Unit Tests
Located in: `./internal/...` and `./cmd/...`
- **Compose Package Tests**: `internal/compose/compose_test.go`
- **Commands Package Tests**: `cmd/commands/save_test.go`

Coverage achieved:
- NewComposeClient: 100%
- load: 100%
- SaveComposeFile: 100%
- Build: 22.6%
- SaveImages: 12.1%
- Save Command: 42.9%

### Integration Tests
Location: `./test/integration/...` (to be created as needed)

### End-to-End Tests
Location: `./test/e2e/...`

## VS Code Integration

### Running Tests
1. **Command Palette** (`Ctrl+Shift+P`):
   - Type "Tasks: Run Task"
   - Select desired test task

2. **Debug Tests**:
   - Set breakpoints in test files
   - Press `F5` and select debug configuration
   - Choose specific test or package to debug

### Coverage Visualization
- Coverage is highlighted directly in the editor
- Green: covered lines
- Red: uncovered lines
- Generate HTML reports with `Go: Generate Coverage Report` task

## Make Targets

```bash
make build                # Build the application
make install              # Install the application
make test                 # Run basic tests
make test-unit           # Run unit tests with race detection
make test-coverage       # Run tests with coverage analysis
make test-race           # Run tests with race detection only
make test-bench          # Run benchmark tests
make test-clean          # Clean test cache
make test-watch          # Watch for changes and run tests
make test-all            # Run comprehensive test suite
make e2e                 # Run end-to-end tests
```

## Coverage Analysis

### Coverage Threshold
- Default threshold: 70%
- Current coverage: ~45%
- Coverage reports generated in HTML format
- Function-level coverage details available

### Improving Coverage
Focus areas for improvement:
- Build function in compose package (currently 22.6%)
- SaveImages function in compose package (currently 12.1%)
- Main application entry points

## Continuous Integration

### GitHub Actions Workflow
- **Triggers**: Push to main/develop, Pull requests
- **Go Versions**: 1.21, 1.22, 1.23, 1.24
- **Steps**:
  1. Checkout code
  2. Set up Go environment
  3. Cache dependencies
  4. Run linting (golangci-lint)
  5. Run tests with coverage
  6. Upload coverage reports
  7. Build application

### Coverage Reporting
- Codecov integration for coverage tracking
- HTML reports uploaded as artifacts
- Coverage badges available for README

## Best Practices

### Writing Tests
1. **Test Naming**: Use descriptive test function names
2. **Table-Driven Tests**: Use for multiple test cases
3. **Mocking**: Mock external dependencies
4. **Coverage**: Aim for high test coverage
5. **Race Detection**: Always test with `-race` flag

### Test Organization
1. **Separate Test Files**: One `*_test.go` file per source file
2. **Helper Functions**: Create reusable test utilities
3. **Setup/Teardown**: Use proper test setup and cleanup
4. **Parallel Tests**: Use `t.Parallel()` for independent tests

### Performance
1. **Short Tests**: Use `-short` flag for quick feedback
2. **Benchmarks**: Write benchmark tests for critical paths
3. **Profiling**: Use Go's built-in profiling tools
4. **Cache Management**: Clean test cache when needed

## Troubleshooting

### Common Issues

1. **Test Cache**: Clean with `make test-clean` or `go clean -testcache`
2. **Race Conditions**: Use `-race` flag to detect data races
3. **Timeouts**: Increase timeout with `-timeout` flag
4. **Dependencies**: Run `go mod download` and `go mod verify`

### Debug Tips

1. **Verbose Output**: Use `-v` flag for detailed test output
2. **Single Test**: Run specific test with `-run TestName`
3. **Coverage**: Use coverage reports to identify untested code
4. **Profiling**: Use `go test -cpuprofile` for performance analysis

## Contributing

When adding new tests:
1. Follow existing test patterns
2. Ensure good coverage of new code
3. Add integration tests for new features
4. Update documentation as needed
5. Verify CI pipeline passes

## Tools and Dependencies

### Required
- Go 1.21+ (tested up to 1.24)
- Make
- Bash (for test script)

### Optional
- fswatch (for file watching)
- bc (for coverage threshold calculations)
- golangci-lint (for advanced linting)
- Codecov account (for coverage reporting)

## Examples

### Running Specific Tests
```bash
# Run only compose tests
go test -v ./internal/compose

# Run specific test function
go test -v -run TestNewComposeClient ./internal/compose

# Run tests with coverage for specific package
go test -cover ./cmd/commands
```

### Debugging Tests
```bash
# Debug with verbose output
go test -v -run TestSaveCmd_Success ./cmd/commands

# Run with race detection
go test -race ./internal/compose

# Generate coverage profile
go test -coverprofile=profile.out ./internal/compose
go tool cover -html=profile.out
```

This comprehensive test configuration provides a solid foundation for maintaining high-quality, well-tested Go code.
