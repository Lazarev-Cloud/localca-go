#!/bin/bash

# LocalCA Integration Test Runner
# This script runs the Next.js integration tests with the actual Go backend

set -e

echo "🚀 Starting LocalCA Integration Tests"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed or not in PATH"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed or not in PATH"
    exit 1
fi

print_status "Checking Go version..."
go version

print_status "Checking Node.js version..."
node --version

print_status "Installing/updating npm dependencies..."
npm install

print_status "Running integration tests..."
echo "This will:"
echo "  1. Build and start the Go backend"
echo "  2. Run Next.js integration tests"
echo "  3. Clean up test environment"
echo ""

# Run the integration tests
if npm run test:integration; then
    print_status "✅ All integration tests passed!"
    exit 0
else
    print_error "❌ Integration tests failed!"
    exit 1
fi 