#!/bin/bash

# LocalCA End-to-End Integration Test Runner
# This script runs comprehensive integration tests with Docker backend

set -e

echo "🚀 Starting LocalCA End-to-End Integration Tests"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to cleanup on exit
cleanup() {
    print_step "Cleaning up test environment..."
    docker-compose -f docker-compose.test.yml down --remove-orphans 2>/dev/null || true
    docker system prune -f 2>/dev/null || true
    rm -rf test-data 2>/dev/null || true
    print_status "Cleanup completed"
}

# Set trap to cleanup on script exit
trap cleanup EXIT

# Check prerequisites
print_step "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH"
    exit 1
fi

if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed or not in PATH"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    print_error "npm is not installed or not in PATH"
    exit 1
fi

print_status "All prerequisites found"

# Display versions
print_step "Environment information:"
echo "Docker version: $(docker --version)"
echo "Docker Compose version: $(docker-compose --version)"
echo "Node.js version: $(node --version)"
echo "npm version: $(npm --version)"

# Install dependencies
print_step "Installing/updating npm dependencies..."
npm install

# Clean up any existing containers
print_step "Cleaning up existing test containers..."
docker-compose -f docker-compose.test.yml down --remove-orphans 2>/dev/null || true

# Remove old test data
print_step "Cleaning up old test data..."
rm -rf test-data 2>/dev/null || true

# Create test data directory
print_step "Setting up test environment..."
mkdir -p test-data
echo "test-ca-password-123" > test-data/cakey.txt

# Start Docker backend
print_step "Starting Docker backend for testing..."
docker-compose -f docker-compose.test.yml up --build -d

# Wait for backend to be ready
print_step "Waiting for backend to be ready..."
BACKEND_URL="http://localhost:8080"
MAX_WAIT=120  # 2 minutes
WAIT_INTERVAL=3

for i in $(seq 1 $MAX_WAIT); do
    if curl -s -f "$BACKEND_URL/api/ca-info" >/dev/null 2>&1; then
        print_status "✅ Backend is ready!"
        break
    fi
    
    if [ $i -eq $MAX_WAIT ]; then
        print_error "❌ Backend failed to start within $MAX_WAIT seconds"
        print_step "Showing backend logs:"
        docker-compose -f docker-compose.test.yml logs backend-test
        exit 1
    fi
    
    echo -n "."
    sleep $WAIT_INTERVAL
done

# Show backend status
print_step "Backend status check:"
SETUP_STATUS=$(curl -s "$BACKEND_URL/api/setup" | jq -r '.success // "unknown"' 2>/dev/null || echo "unknown")
CA_STATUS=$(curl -s "$BACKEND_URL/api/ca-info" | jq -r '.success // "unknown"' 2>/dev/null || echo "unknown")

echo "Setup endpoint: $SETUP_STATUS"
echo "CA info endpoint: $CA_STATUS"

# Run the integration tests
print_step "Running integration tests..."
echo "This will test:"
echo "  ✓ Backend API endpoints"
echo "  ✓ Frontend component rendering"
echo "  ✓ Setup and login workflows"
echo "  ✓ Error handling"
echo "  ✓ CORS configuration"
echo ""

# Run tests with detailed output
if npm run test:integration -- --verbose --detectOpenHandles; then
    print_status "✅ All integration tests passed!"
    
    # Run additional manual verification
    print_step "Running manual verification checks..."
    
    # Test setup endpoint
    print_step "Testing setup endpoint..."
    SETUP_RESPONSE=$(curl -s "$BACKEND_URL/api/setup")
    echo "Setup response: $SETUP_RESPONSE"
    
    # Test CA info endpoint
    print_step "Testing CA info endpoint..."
    CA_RESPONSE=$(curl -s "$BACKEND_URL/api/ca-info")
    echo "CA info response: $CA_RESPONSE"
    
    # Test login endpoint
    print_step "Testing login endpoint..."
    LOGIN_RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"wrongpass"}')
    echo "Login response: $LOGIN_RESPONSE"
    
    print_status "🎉 All tests completed successfully!"
    
else
    print_error "❌ Integration tests failed!"
    
    print_step "Showing backend logs for debugging:"
    docker-compose -f docker-compose.test.yml logs backend-test
    
    exit 1
fi 