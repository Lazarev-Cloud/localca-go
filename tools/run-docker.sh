#!/bin/bash

# Exit on error
set -e

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting LocalCA Docker Deployment${NC}"

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: docker-compose is not installed.${NC}"
    echo -e "Please install docker-compose before running this script."
    exit 1
fi

# Check if data directory exists, if not create it
if [ ! -d "./data" ]; then
    echo -e "${YELLOW}Creating data directory...${NC}"
    mkdir -p ./data
fi

# Check if cakey.txt exists, if not create it
if [ ! -f "./data/cakey.txt" ]; then
    echo -e "${YELLOW}Creating cakey.txt with random password...${NC}"
    # Generate a random password
    openssl rand -base64 16 > ./cakey.txt
    # Make sure it's moved to the right location
    cp ./cakey.txt ./data/cakey.txt
fi

# Build and start the containers
echo -e "${GREEN}Building and starting Docker containers...${NC}"
docker-compose down || true
docker-compose build
docker-compose up -d

echo -e "${GREEN}Docker containers are up and running!${NC}"
echo -e "- Frontend UI: http://localhost:3000"
echo -e "- Backend API: http://localhost:8080"
echo ""
echo -e "${YELLOW}Important Notes:${NC}"
echo -e "1. On first run, you'll need to complete the setup at http://localhost:3000/setup"
echo -e "2. The initial setup token can be found in the logs:"
echo -e "   ${YELLOW}docker-compose logs backend | grep 'Setup token'${NC}"
echo ""
echo -e "${GREEN}To stop the services, run:${NC}"
echo -e "docker-compose down"
echo ""

# Show logs after startup
echo -e "${GREEN}Showing startup logs:${NC}"
docker-compose logs --tail=20

echo ""
echo -e "${GREEN}Done!${NC}" 