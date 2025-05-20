#!/bin/bash

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running LocalCA-Go Tests${NC}"
echo "==============================="

# Run tests with coverage
echo -e "${BLUE}Running tests with coverage...${NC}"
go test -v -cover ./pkg/... 

# Check if tests passed
if [ $? -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Tests failed!${NC}"
    exit 1
fi 