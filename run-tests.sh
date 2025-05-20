#!/bin/bash

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running LocalCA-Go Tests${NC}"
echo "==============================="

# Function to run tests with coverage
run_tests() {
    echo -e "${BLUE}Running tests for $1...${NC}"
    go test -v -cover $1
    return $?
}

# Run all package tests
echo -e "${BLUE}Running all package tests...${NC}"
run_tests "./pkg/..."
if [ $? -ne 0 ]; then
    echo -e "${RED}Package tests failed!${NC}"
    exit 1
fi

# Run main package test
echo -e "${BLUE}Running main package test...${NC}"
run_tests "./main_test.go"
if [ $? -ne 0 ]; then
    echo -e "${RED}Main package test failed!${NC}"
    exit 1
fi

# Check if Docker is available
echo -e "${BLUE}Checking Docker availability...${NC}"
if command -v docker &> /dev/null; then
    echo -e "${GREEN}Docker is available.${NC}"
    
    # Check if Docker is running
    if docker info &> /dev/null; then
        echo -e "${GREEN}Docker is running.${NC}"
        
        # Test Docker build
        echo -e "${BLUE}Testing Docker build...${NC}"
        docker build -t localca-go-backend:test -f Dockerfile . 
        if [ $? -ne 0 ]; then
            echo -e "${RED}Docker build failed!${NC}"
            exit 1
        else
            echo -e "${GREEN}Docker build successful.${NC}"
        fi
        
        # Test Docker Compose if available
        if command -v docker-compose &> /dev/null; then
            echo -e "${BLUE}Testing Docker Compose configuration...${NC}"
            docker-compose config
            if [ $? -ne 0 ]; then
                echo -e "${RED}Docker Compose configuration failed!${NC}"
                exit 1
            else
                echo -e "${GREEN}Docker Compose configuration is valid.${NC}"
            fi
        else
            echo -e "${YELLOW}Docker Compose not available, skipping Docker Compose tests.${NC}"
        fi
    else
        echo -e "${YELLOW}Docker is not running, skipping Docker tests.${NC}"
    fi
else
    echo -e "${YELLOW}Docker not available, skipping Docker tests.${NC}"
fi

echo -e "${GREEN}All tests passed!${NC}"
exit 0 