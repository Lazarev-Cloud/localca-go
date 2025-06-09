#!/bin/bash

echo "=== LocalCA Comprehensive Test Suite ==="
echo "Running all tests and quality checks..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Function to print section header
print_header() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Track overall success
OVERALL_SUCCESS=0

print_header "Go Module and Dependencies"
go mod tidy
print_status $? "Go module cleanup"

go mod verify
print_status $? "Go module verification"

print_header "Go Build Test"
go build -o /tmp/localca-test ./...
BUILD_STATUS=$?
print_status $BUILD_STATUS "Go build"
if [ $BUILD_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Go Linting (go vet)"
go vet ./...
VET_STATUS=$?
print_status $VET_STATUS "Go vet"
if [ $VET_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Go Tests"
go test -v ./...
GO_TEST_STATUS=$?
print_status $GO_TEST_STATUS "Go tests"
if [ $GO_TEST_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Go Test Coverage"
go test -coverprofile=coverage.out ./...
COVERAGE_STATUS=$?
if [ $COVERAGE_STATUS -eq 0 ]; then
    go tool cover -html=coverage.out -o coverage.html
    COVERAGE_PERCENT=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo -e "${GREEN}✓ Coverage report generated: $COVERAGE_PERCENT${NC}"
else
    print_status $COVERAGE_STATUS "Go test coverage"
    OVERALL_SUCCESS=1
fi

print_header "Next.js Dependencies"
npm install
NPM_INSTALL_STATUS=$?
print_status $NPM_INSTALL_STATUS "npm install"
if [ $NPM_INSTALL_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Next.js Linting"
npm run lint
LINT_STATUS=$?
print_status $LINT_STATUS "Next.js lint"
if [ $LINT_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Next.js Build Test"
npm run build
BUILD_NEXT_STATUS=$?
print_status $BUILD_NEXT_STATUS "Next.js build"
if [ $BUILD_NEXT_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Next.js Tests"
npm run test:ci
JEST_STATUS=$?
print_status $JEST_STATUS "Jest tests"
if [ $JEST_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "TypeScript Type Checking"
npx tsc --noEmit
TSC_STATUS=$?
print_status $TSC_STATUS "TypeScript compilation"
if [ $TSC_STATUS -ne 0 ]; then
    OVERALL_SUCCESS=1
fi

print_header "Test Summary"
if [ $OVERALL_SUCCESS -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed successfully!${NC}"
    echo -e "${GREEN}✓ Go build and tests${NC}"
    echo -e "${GREEN}✓ Next.js build and tests${NC}"
    echo -e "${GREEN}✓ Linting and type checking${NC}"
    echo -e "${GREEN}✓ Code coverage generated${NC}"
else
    echo -e "${RED}❌ Some tests failed. Please check the output above.${NC}"
fi

echo
echo "=== Test Results ==="
echo "Go Tests: $([ $GO_TEST_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "Go Build: $([ $BUILD_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "Go Vet: $([ $VET_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "Next.js Tests: $([ $JEST_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "Next.js Build: $([ $BUILD_NEXT_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "Next.js Lint: $([ $LINT_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"
echo "TypeScript: $([ $TSC_STATUS -eq 0 ] && echo 'PASS' || echo 'FAIL')"

exit $OVERALL_SUCCESS 