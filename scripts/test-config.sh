#!/bin/bash

# Go Test Configuration Script
# This script provides various testing utilities for the docker-deliver project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TIMEOUT="5m"
COVERAGE_THRESHOLD=70
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
VERBOSE=false
RACE=true
SHORT=false

# Help function
show_help() {
    echo "Go Test Configuration Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  unit            Run unit tests only"
    echo "  integration     Run integration tests only"
    echo "  e2e             Run end-to-end tests only"
    echo "  all             Run all tests"
    echo "  coverage        Run tests with coverage report"
    echo "  benchmark       Run benchmark tests"
    echo "  clean           Clean test cache"
    echo "  watch           Watch for changes and run tests"
    echo ""
    echo "Options:"
    echo "  -v, --verbose   Enable verbose output"
    echo "  -t, --timeout   Set test timeout (default: 5m)"
    echo "  -s, --short     Run short tests only"
    echo "  --no-race       Disable race detection"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 unit -v"
    echo "  $0 coverage --timeout 10m"
    echo "  $0 all --short"
}

# Parse command line arguments
COMMAND=""
while [[ $# -gt 0 ]]; do
    case $1 in
        unit|integration|e2e|all|coverage|benchmark|clean|watch)
            COMMAND="$1"
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        --no-race)
            RACE=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Build test flags
build_test_flags() {
    local flags=()
    
    if [ "$VERBOSE" = true ]; then
        flags+=("-v")
    fi
    
    if [ "$RACE" = true ]; then
        flags+=("-race")
    fi
    
    if [ "$SHORT" = true ]; then
        flags+=("-short")
    fi
    
    flags+=("-timeout=$TIMEOUT")
    
    echo "${flags[@]}"
}

# Print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Run unit tests
run_unit_tests() {
    print_status $BLUE "Running unit tests..."
    local flags=($(build_test_flags))
    go test "${flags[@]}" ./internal/... ./cmd/...
}

# Run integration tests
run_integration_tests() {
    print_status $BLUE "Running integration tests..."
    local flags=($(build_test_flags))
    if [ -d "./test/integration" ]; then
        go test "${flags[@]}" ./test/integration/...
    else
        print_status $YELLOW "No integration tests found"
    fi
}

# Run e2e tests
run_e2e_tests() {
    print_status $BLUE "Running end-to-end tests..."
    local flags=($(build_test_flags))
    if [ -d "./test/e2e" ]; then
        go test "${flags[@]}" ./test/e2e/...
    else
        print_status $YELLOW "No e2e tests found"
    fi
}

# Run all tests
run_all_tests() {
    print_status $BLUE "Running all tests..."
    local flags=($(build_test_flags))
    go test "${flags[@]}" ./...
}

# Run tests with coverage
run_coverage_tests() {
    print_status $BLUE "Running tests with coverage..."
    local flags=($(build_test_flags))
    
    # Run tests with coverage
    go test "${flags[@]}" -coverprofile="$COVERAGE_FILE" -covermode=atomic ./internal/... ./cmd/...
    
    # Generate HTML report
    go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
    print_status $GREEN "Coverage report generated: $COVERAGE_HTML"
    
    # Show coverage summary
    go tool cover -func="$COVERAGE_FILE"
    
    # Check coverage threshold
    local coverage_percent=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
        print_status $GREEN "Coverage threshold met: ${coverage_percent}% >= ${COVERAGE_THRESHOLD}%"
    else
        print_status $RED "Coverage threshold not met: ${coverage_percent}% < ${COVERAGE_THRESHOLD}%"
        exit 1
    fi
}

# Run benchmark tests
run_benchmark_tests() {
    print_status $BLUE "Running benchmark tests..."
    go test -bench=. -benchmem ./...
}

# Clean test cache
clean_test_cache() {
    print_status $BLUE "Cleaning test cache..."
    go clean -testcache
    print_status $GREEN "Test cache cleaned"
}

# Watch for changes and run tests
watch_tests() {
    print_status $BLUE "Watching for changes..."
    
    if ! command -v fswatch &> /dev/null; then
        print_status $RED "fswatch not installed. Please install it first:"
        print_status $YELLOW "  brew install fswatch  # macOS"
        print_status $YELLOW "  apt-get install fswatch  # Ubuntu"
        exit 1
    fi
    
    fswatch -o . --exclude=".*" --include="\.go$" | while read num; do
        print_status $YELLOW "Changes detected, running tests..."
        run_unit_tests
    done
}

# Main execution
main() {
    case $COMMAND in
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        e2e)
            run_e2e_tests
            ;;
        all)
            run_all_tests
            ;;
        coverage)
            run_coverage_tests
            ;;
        benchmark)
            run_benchmark_tests
            ;;
        clean)
            clean_test_cache
            ;;
        watch)
            watch_tests
            ;;
        "")
            print_status $RED "No command specified"
            show_help
            exit 1
            ;;
        *)
            print_status $RED "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Check if go is installed
if ! command -v go &> /dev/null; then
    print_status $RED "Go is not installed or not in PATH"
    exit 1
fi

# Check if bc is available for coverage threshold comparison
if ! command -v bc &> /dev/null && [ "$COMMAND" = "coverage" ]; then
    print_status $YELLOW "bc not installed, skipping coverage threshold check"
fi

main
