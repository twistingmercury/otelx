#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_ROOT="${PROJ_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"

# sources
source "${PROJ_ROOT}/scripts/print.sh"

# global var declarations
COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"

# internal functions

build::test() {
    print::header "Running Tests"
    print::step "1" "Running go test with race detection"

    cd "${PROJ_ROOT}"
    if ! go test -race -v ./...; then
        print::error "Tests failed"
        return 1
    fi

    print::success "All tests passed"
    return 0
}

build::compile() {
    print::header "Building Library"
    print::step "1" "Running go build"

    cd "${PROJ_ROOT}"
    if ! go build ./...; then
        print::error "Build failed"
        return 1
    fi

    print::success "Build completed successfully"
    return 0
}

build::vet() {
    print::header "Running go vet"

    cd "${PROJ_ROOT}"
    if ! go vet ./...; then
        print::error "go vet found issues"
        return 1
    fi

    print::success "go vet passed"
    return 0
}

build::staticcheck() {
    print::header "Running staticcheck"

    cd "${PROJ_ROOT}"
    if ! command -v staticcheck > /dev/null 2>&1; then
        print::warning "staticcheck not installed, skipping"
        print::info "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
        return 0
    fi

    if ! staticcheck ./...; then
        print::error "staticcheck found issues"
        return 1
    fi

    print::success "staticcheck passed"
    return 0
}

build::lint() {
    print::header "Running All Linters"
    local lint_failed=0

    print::step "1/2" "go vet"
    if ! build::vet; then
        lint_failed=1
    fi

    print::step "2/2" "staticcheck"
    if ! build::staticcheck; then
        lint_failed=1
    fi

    if [ "${lint_failed}" -eq 1 ]; then
        print::error "Linting failed"
        return 1
    fi

    print::success "All linters passed"
    return 0
}

build::coverage() {
    print::header "Running Tests with Coverage"
    print::step "1" "Generating coverage report"

    cd "${PROJ_ROOT}"
    if ! go test -race -coverprofile="${COVERAGE_FILE}" -covermode=atomic ./...; then
        print::error "Tests failed"
        return 1
    fi

    print::step "2" "Coverage summary"
    go tool cover -func="${COVERAGE_FILE}"

    print::success "Coverage report generated: ${COVERAGE_FILE}"
    print::info "View HTML report with: go tool cover -html=${COVERAGE_FILE}"
    return 0
}

build::all() {
    print::header "Running Full Build Pipeline"

    print::step "1/3" "Lint"
    if ! build::lint; then
        return 1
    fi

    print::step "2/3" "Test"
    if ! build::test; then
        return 1
    fi

    print::step "3/3" "Build"
    if ! build::compile; then
        return 1
    fi

    print::success "Full build pipeline completed successfully"
    return 0
}

build::help() {
    print::header "otelx Build Commands"
    printf "Usage: make <target>\n\n"
    printf "Targets:\n"
    printf "  test        Run tests with race detection\n"
    printf "  build       Build the library\n"
    printf "  vet         Run go vet\n"
    printf "  staticcheck Run staticcheck\n"
    printf "  lint        Run all linters (vet + staticcheck)\n"
    printf "  coverage    Run tests with coverage report\n"
    printf "  all         Run lint, test, and build\n"
    printf "  help        Show this help message\n"
    printf "\n"
}
