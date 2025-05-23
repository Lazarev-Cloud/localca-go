#!/bin/bash

# Quick rebuild and test script
# This script rebuilds the application and runs basic tests

echo "🔧 LocalCA Quick Rebuild & Test"
echo "================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check command success
check_command() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $1 successful${NC}"
    else
        echo -e "${RED}❌ $1 failed${NC}"
        exit 1
    fi
}

# Build backend
echo -e "${YELLOW}🏗️  Building backend...${NC}"
go build -o localca-go
check_command "Backend build"

# Install frontend dependencies (if needed)
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
    npm install
    check_command "Frontend dependencies installation"
fi

# Build frontend
echo -e "${YELLOW}🎨 Building frontend...${NC}"
npm run build
check_command "Frontend build"

# Run basic tests
echo -e "${YELLOW}🧪 Running Go tests...${NC}"
go test ./pkg/... -v
check_command "Go tests"

# Check if Docker is available and build containers
if command -v docker &> /dev/null; then
    echo -e "${YELLOW}🐳 Building Docker containers...${NC}"
    docker build -t localca-backend -f Dockerfile .
    check_command "Backend Docker build"
    
    docker build -t localca-frontend -f Dockerfile.frontend .
    check_command "Frontend Docker build"
else
    echo -e "${YELLOW}⚠️  Docker not available, skipping container builds${NC}"
fi

# Test if the binary runs
echo -e "${YELLOW}🚀 Testing binary execution...${NC}"
timeout 5s ./localca-go &
BACKEND_PID=$!
sleep 2

# Check if the process is running
if kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${GREEN}✅ Backend starts successfully${NC}"
    kill $BACKEND_PID
else
    echo -e "${RED}❌ Backend failed to start${NC}"
fi

echo -e "\n${GREEN}🎉 Quick rebuild completed successfully!${NC}"
echo -e "${YELLOW}📋 Summary:${NC}"
echo "   - Backend binary: ./localca-go"
echo "   - Frontend build: .next/"
echo "   - Docker images: localca-backend, localca-frontend"
echo ""
echo -e "${YELLOW}🚀 Next steps:${NC}"
echo "   - Start with Docker: docker-compose up -d"
echo "   - Or run standalone: ./localca-go & npm start"
echo "   - Access at: http://localhost:3000" 