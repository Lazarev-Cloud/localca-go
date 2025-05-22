#!/bin/bash
set -e

# Ensure scripts directory exists
mkdir -p scripts
chmod +x scripts/verify-build.sh 2>/dev/null || true

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}===== LocalCA Docker Tests =====${NC}"

# Clean up any existing test containers
echo -e "${YELLOW}Cleaning up previous test containers...${NC}"
docker-compose -f ../build/docker-compose.test.yml down -v 2>/dev/null || true

# Build containers first to avoid long waits during testing
echo -e "${YELLOW}Building test containers (this may take a minute)...${NC}"
docker-compose -f ../build/docker-compose.test.yml build
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Test containers built successfully${NC}"
else
    echo -e "${RED}❌ Building test containers failed${NC}"
    exit 1
fi

# Run backend tests with verbose output
echo -e "${YELLOW}Running backend tests (this may take a minute)...${NC}"
docker-compose -f ../build/docker-compose.test.yml run --rm backend-test
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Backend tests passed${NC}"
else
    echo -e "${RED}❌ Backend tests failed${NC}"
    exit 1
fi

# Run frontend tests with verbose output
echo -e "${YELLOW}Running frontend tests...${NC}"
docker-compose -f ../build/docker-compose.test.yml run --rm frontend-test
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Frontend tests passed${NC}"
else
    echo -e "${RED}❌ Frontend tests failed${NC}"
    exit 1
fi

# Verify build with more verbose output 
echo -e "${YELLOW}Verifying build process...${NC}"
docker-compose -f ../build/docker-compose.test.yml run --rm build-check
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build verification passed${NC}"
else
    echo -e "${RED}❌ Build verification failed${NC}"
    exit 1
fi

echo -e "${GREEN}===== All tests passed! =====${NC}"

# Clean up test containers and volumes
echo -e "${YELLOW}Cleaning up test containers and volumes...${NC}"
docker-compose -f ../build/docker-compose.test.yml down -v

exit 0 