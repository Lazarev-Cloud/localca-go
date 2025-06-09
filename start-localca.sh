#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================"
echo -e "   LocalCA Docker Deployment Script"
echo -e "========================================${NC}"
echo

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}[ERROR] Docker is not running or not installed.${NC}"
    echo "Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}[ERROR] docker-compose is not installed.${NC}"
    echo "Please install Docker Compose and try again."
    exit 1
fi

echo -e "${GREEN}[INFO] Docker is running and docker-compose is available.${NC}"
echo

# Create data directory if it doesn't exist
if [ ! -d "data" ]; then
    echo -e "${YELLOW}[INFO] Creating data directory...${NC}"
    mkdir -p data
fi

# Create cakey.txt if it doesn't exist
if [ ! -f "data/cakey.txt" ]; then
    echo -e "${YELLOW}[INFO] Creating CA key password file...${NC}"
    echo "LocalCA_SecurePassword_2024!" > data/cakey.txt
fi

echo -e "${YELLOW}[INFO] Stopping any existing containers...${NC}"
docker-compose down >/dev/null 2>&1

echo -e "${YELLOW}[INFO] Building Docker images...${NC}"
docker-compose build

echo -e "${YELLOW}[INFO] Starting LocalCA services...${NC}"
docker-compose up -d

echo
echo -e "${GREEN}[SUCCESS] LocalCA is starting up!${NC}"
echo
echo -e "${BLUE}Services:${NC}"
echo "- Frontend UI:     http://localhost:3000"
echo "- Backend API:     http://localhost:8080"
echo "- ACME Server:     http://localhost:8555"
echo "- PostgreSQL:      localhost:5432"
echo "- MinIO Console:   http://localhost:9001"
echo "- KeyDB:           localhost:6379"
echo
echo -e "${BLUE}Credentials:${NC}"
echo "- Database: localca / localca_postgres_password"
echo "- MinIO:    localca / localca_minio_password"
echo "- KeyDB:    localca_keydb_password"
echo

echo -e "${YELLOW}[INFO] Waiting for services to start...${NC}"
sleep 10

echo -e "${YELLOW}[INFO] Checking service status...${NC}"
docker-compose ps

echo
echo -e "${YELLOW}[INFO] Getting setup token from backend logs...${NC}"
SETUP_TOKEN=$(docker-compose logs backend 2>/dev/null | grep "Setup token" | tail -1)
if [ -n "$SETUP_TOKEN" ]; then
    echo -e "${GREEN}$SETUP_TOKEN${NC}"
else
    echo -e "${YELLOW}[INFO] Setup token not found yet. Services may still be starting.${NC}"
    echo "Check logs with: docker-compose logs backend"
fi

echo
echo -e "${BLUE}[INFO] To complete setup:${NC}"
echo "1. Visit http://localhost:3000/setup"
echo "2. Use the setup token from the logs above"
echo "3. Create your admin account"
echo
echo -e "${BLUE}[INFO] Useful commands:${NC}"
echo "- Stop services: docker-compose down"
echo "- View logs: docker-compose logs [service-name]"
echo "- Restart service: docker-compose restart [service-name]"
echo
echo -e "${YELLOW}Press Enter to show live logs (Ctrl+C to exit)...${NC}"
read

docker-compose logs -f 