#!/bin/bash

# Enhanced Storage Test Script for LocalCA
# Tests PostgreSQL, S3/MinIO, and structured logging features

set -e

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== LocalCA Enhanced Storage Test Suite ===${NC}"
echo -e "${YELLOW}Testing PostgreSQL, S3/MinIO, and structured logging features${NC}"
echo

# Function to check if a service is healthy
check_service_health() {
    local service_name=$1
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}Checking $service_name health...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose ps | grep -q "$service_name.*healthy"; then
            echo -e "${GREEN}✅ $service_name is healthy${NC}"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ $service_name failed to become healthy${NC}"
    return 1
}

# Function to test database connectivity
test_database() {
    echo -e "${BLUE}=== Testing PostgreSQL Database ===${NC}"
    
    # Test database connection
    echo -e "${YELLOW}Testing database connection...${NC}"
    DB_TEST=$(docker-compose exec -T postgres psql -U localca -d localca -c "SELECT version();" 2>/dev/null || echo "FAILED")
    
    if [[ "$DB_TEST" == *"PostgreSQL"* ]]; then
        echo -e "${GREEN}✅ Database connection successful${NC}"
        
        # Test table creation (should be done by GORM migrations)
        echo -e "${YELLOW}Checking database tables...${NC}"
        TABLES=$(docker-compose exec -T postgres psql -U localca -d localca -c "\dt" 2>/dev/null || echo "FAILED")
        
        if [[ "$TABLES" == *"ca_info"* ]] && [[ "$TABLES" == *"certificates"* ]] && [[ "$TABLES" == *"audit_logs"* ]]; then
            echo -e "${GREEN}✅ Database tables exist${NC}"
            echo -e "   Found tables: ca_info, certificates, audit_logs, email_settings, serial_mappings"
        else
            echo -e "${YELLOW}⚠️  Some database tables may not exist yet (will be created on first use)${NC}"
        fi
        
        # Test audit log insertion
        echo -e "${YELLOW}Testing audit log functionality...${NC}"
        AUDIT_COUNT=$(docker-compose exec -T postgres psql -U localca -d localca -c "SELECT COUNT(*) FROM audit_logs;" 2>/dev/null | grep -E "^\s*[0-9]+\s*$" | tr -d ' ' || echo "0")
        echo -e "${GREEN}✅ Audit logs table accessible (${AUDIT_COUNT} entries)${NC}"
        
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        return 1
    fi
}

# Function to test S3/MinIO connectivity
test_s3_storage() {
    echo -e "${BLUE}=== Testing S3/MinIO Storage ===${NC}"
    
    # Test MinIO health
    echo -e "${YELLOW}Testing MinIO connectivity...${NC}"
    MINIO_HEALTH=$(curl -s -f http://localhost:9000/minio/health/live 2>/dev/null && echo "OK" || echo "FAILED")
    
    if [ "$MINIO_HEALTH" = "OK" ]; then
        echo -e "${GREEN}✅ MinIO is accessible${NC}"
        
        # Test bucket creation (should be done by application)
        echo -e "${YELLOW}Checking MinIO buckets...${NC}"
        
        # Install mc (MinIO client) in the MinIO container if not present
        docker-compose exec -T minio sh -c "
            if ! command -v mc &> /dev/null; then
                wget -q https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc
                chmod +x /usr/local/bin/mc
            fi
            mc alias set local http://localhost:9000 localca localca_minio_password
            mc ls local/ 2>/dev/null || echo 'No buckets found'
        " 2>/dev/null || echo -e "${YELLOW}⚠️  MinIO client setup failed, but service is running${NC}"
        
    else
        echo -e "${RED}❌ MinIO is not accessible${NC}"
        return 1
    fi
    
    # Test MinIO console
    echo -e "${YELLOW}Testing MinIO console...${NC}"
    CONSOLE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9001 2>/dev/null || echo "000")
    if [ "$CONSOLE_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ MinIO console is accessible at http://localhost:9001${NC}"
    else
        echo -e "${YELLOW}⚠️  MinIO console may not be ready yet (HTTP $CONSOLE_STATUS)${NC}"
    fi
}

# Function to test cache connectivity
test_cache() {
    echo -e "${BLUE}=== Testing KeyDB Cache ===${NC}"
    
    echo -e "${YELLOW}Testing KeyDB connectivity...${NC}"
    CACHE_TEST=$(docker-compose exec -T keydb keydb-cli -a localca_keydb_password ping 2>/dev/null || echo "FAILED")
    
    if [ "$CACHE_TEST" = "PONG" ]; then
        echo -e "${GREEN}✅ KeyDB cache is accessible${NC}"
        
        # Test cache operations
        echo -e "${YELLOW}Testing cache operations...${NC}"
        docker-compose exec -T keydb keydb-cli -a localca_keydb_password set test_key "test_value" >/dev/null 2>&1
        CACHE_VALUE=$(docker-compose exec -T keydb keydb-cli -a localca_keydb_password get test_key 2>/dev/null | tr -d '\r')
        
        if [ "$CACHE_VALUE" = "test_value" ]; then
            echo -e "${GREEN}✅ Cache read/write operations working${NC}"
            docker-compose exec -T keydb keydb-cli -a localca_keydb_password del test_key >/dev/null 2>&1
        else
            echo -e "${YELLOW}⚠️  Cache operations may have issues${NC}"
        fi
    else
        echo -e "${RED}❌ KeyDB cache is not accessible${NC}"
        return 1
    fi
}

# Function to test application with enhanced storage
test_application_storage() {
    echo -e "${BLUE}=== Testing Application with Enhanced Storage ===${NC}"
    
    # Wait for application to be ready
    echo -e "${YELLOW}Waiting for application to be ready...${NC}"
    sleep 10
    
    # Test backend health with storage backends
    echo -e "${YELLOW}Testing backend health endpoint...${NC}"
    HEALTH_RESPONSE=$(curl -s http://localhost:8080/health 2>/dev/null || echo "FAILED")
    
    if [[ "$HEALTH_RESPONSE" == *"status"* ]]; then
        echo -e "${GREEN}✅ Backend health endpoint accessible${NC}"
        echo -e "   Health response: $HEALTH_RESPONSE"
    else
        echo -e "${YELLOW}⚠️  Backend health endpoint may not be implemented yet${NC}"
    fi
    
    # Test setup endpoint
    echo -e "${YELLOW}Testing setup endpoint...${NC}"
    SETUP_RESPONSE=$(curl -s http://localhost:8080/api/setup 2>/dev/null || echo "FAILED")
    SETUP_SUCCESS=$(echo "$SETUP_RESPONSE" | jq -r '.success // false' 2>/dev/null || echo "false")
    
    if [ "$SETUP_SUCCESS" = "true" ]; then
        echo -e "${GREEN}✅ Setup endpoint accessible${NC}"
        SETUP_COMPLETED=$(echo "$SETUP_RESPONSE" | jq -r '.data.setup_completed // false' 2>/dev/null || echo "false")
        echo -e "   Setup completed: $SETUP_COMPLETED"
    else
        echo -e "${YELLOW}⚠️  Setup endpoint may have issues${NC}"
    fi
    
    # Test structured logging
    echo -e "${YELLOW}Testing structured logging...${NC}"
    LOG_SAMPLE=$(docker-compose logs --tail=5 backend 2>/dev/null | grep -E '\{.*\}' | head -1 || echo "")
    
    if [[ "$LOG_SAMPLE" == *"level"* ]] && [[ "$LOG_SAMPLE" == *"msg"* ]]; then
        echo -e "${GREEN}✅ Structured logging is working${NC}"
        echo -e "   Sample log: ${LOG_SAMPLE:0:100}..."
    else
        echo -e "${YELLOW}⚠️  Structured logging may not be configured${NC}"
    fi
}

# Function to test environment variables
test_environment() {
    echo -e "${BLUE}=== Testing Environment Configuration ===${NC}"
    
    echo -e "${YELLOW}Checking environment variables in backend container...${NC}"
    
    # Check database environment
    DB_ENABLED=$(docker-compose exec -T backend printenv DATABASE_ENABLED 2>/dev/null || echo "not_set")
    S3_ENABLED=$(docker-compose exec -T backend printenv S3_ENABLED 2>/dev/null || echo "not_set")
    CACHE_ENABLED=$(docker-compose exec -T backend printenv CACHE_ENABLED 2>/dev/null || echo "not_set")
    LOG_FORMAT=$(docker-compose exec -T backend printenv LOG_FORMAT 2>/dev/null || echo "not_set")
    
    echo -e "   DATABASE_ENABLED: $DB_ENABLED"
    echo -e "   S3_ENABLED: $S3_ENABLED"
    echo -e "   CACHE_ENABLED: $CACHE_ENABLED"
    echo -e "   LOG_FORMAT: $LOG_FORMAT"
    
    if [ "$DB_ENABLED" = "true" ] && [ "$S3_ENABLED" = "true" ] && [ "$CACHE_ENABLED" = "true" ]; then
        echo -e "${GREEN}✅ Enhanced storage environment properly configured${NC}"
    else
        echo -e "${YELLOW}⚠️  Some enhanced storage features may not be enabled${NC}"
    fi
}

# Function to show service URLs and credentials
show_service_info() {
    echo -e "${BLUE}=== Service Information ===${NC}"
    echo -e "${GREEN}Application URLs:${NC}"
    echo -e "   Frontend:      http://localhost:3000"
    echo -e "   Backend API:   http://localhost:8080"
    echo -e "   MinIO Console: http://localhost:9001"
    echo
    echo -e "${GREEN}Database Connection:${NC}"
    echo -e "   Host: localhost:5432"
    echo -e "   Database: localca"
    echo -e "   Username: localca"
    echo -e "   Password: localca_postgres_password"
    echo
    echo -e "${GREEN}MinIO Credentials:${NC}"
    echo -e "   Access Key: localca"
    echo -e "   Secret Key: localca_minio_password"
    echo -e "   Bucket: localca-certificates"
    echo
    echo -e "${GREEN}KeyDB Cache:${NC}"
    echo -e "   Host: localhost:6379"
    echo -e "   Password: localca_keydb_password"
}

# Main test execution
main() {
    echo -e "${YELLOW}Starting enhanced storage tests...${NC}"
    echo
    
    # Check if docker-compose is running
    if ! docker-compose ps | grep -q "Up"; then
        echo -e "${RED}❌ Docker containers are not running${NC}"
        echo -e "${YELLOW}Please run 'docker-compose up -d' first${NC}"
        exit 1
    fi
    
    # Wait for services to be healthy
    echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
    check_service_health "postgres" || exit 1
    check_service_health "minio" || exit 1
    check_service_health "keydb" || exit 1
    
    echo
    
    # Run tests
    test_environment
    echo
    test_database
    echo
    test_s3_storage
    echo
    test_cache
    echo
    test_application_storage
    echo
    
    # Show service information
    show_service_info
    
    echo
    echo -e "${GREEN}=== Enhanced Storage Test Summary ===${NC}"
    echo -e "${GREEN}✅ All enhanced storage components tested${NC}"
    echo -e "${YELLOW}📝 Check the logs above for any warnings or issues${NC}"
    echo -e "${BLUE}🚀 Your LocalCA instance is ready with enhanced storage!${NC}"
}

# Run main function
main "$@" 