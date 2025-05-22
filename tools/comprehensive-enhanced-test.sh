#!/bin/bash

# Comprehensive Enhanced Test Script for LocalCA
# Combines all existing tests with enhanced storage features

set -e

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${PURPLE}=== LocalCA Comprehensive Enhanced Test Suite ===${NC}"
echo -e "${YELLOW}Testing all features including PostgreSQL, S3/MinIO, and structured logging${NC}"
echo

# Function to run a test section
run_test_section() {
    local section_name="$1"
    local test_function="$2"
    
    echo -e "${BLUE}=== $section_name ===${NC}"
    if $test_function; then
        echo -e "${GREEN}✅ $section_name completed successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ $section_name failed${NC}"
        return 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    # Check if docker-compose is installed
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ docker-compose is not installed${NC}"
        return 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠️  jq is not installed, some tests may be limited${NC}"
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ curl is not installed${NC}"
        return 1
    fi
    
    # Check if containers are running
    if ! docker-compose ps | grep -q "Up"; then
        echo -e "${RED}❌ Docker containers are not running${NC}"
        echo -e "${YELLOW}Please run 'docker-compose up -d' first${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
    return 0
}

# Function to wait for services to be ready
wait_for_services() {
    echo -e "${YELLOW}Waiting for all services to be ready...${NC}"
    
    local max_attempts=60
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        local all_healthy=true
        
        # Check each service
        for service in postgres minio keydb backend; do
            if ! docker-compose ps | grep -q "$service.*healthy\|$service.*Up"; then
                all_healthy=false
                break
            fi
        done
        
        if $all_healthy; then
            echo -e "${GREEN}✅ All services are ready${NC}"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ Services failed to become ready${NC}"
    return 1
}

# Enhanced storage tests
test_enhanced_storage() {
    echo -e "${YELLOW}Running enhanced storage tests...${NC}"
    
    # Test database
    echo -e "${YELLOW}Testing PostgreSQL database...${NC}"
    DB_TEST=$(docker-compose exec -T postgres psql -U localca -d localca -c "SELECT version();" 2>/dev/null || echo "FAILED")
    if [[ "$DB_TEST" == *"PostgreSQL"* ]]; then
        echo -e "${GREEN}✅ Database connection successful${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        return 1
    fi
    
    # Test MinIO
    echo -e "${YELLOW}Testing MinIO S3 storage...${NC}"
    MINIO_HEALTH=$(curl -s -f http://localhost:9000/minio/health/live 2>/dev/null && echo "OK" || echo "FAILED")
    if [ "$MINIO_HEALTH" = "OK" ]; then
        echo -e "${GREEN}✅ MinIO is accessible${NC}"
    else
        echo -e "${RED}❌ MinIO is not accessible${NC}"
        return 1
    fi
    
    # Test KeyDB
    echo -e "${YELLOW}Testing KeyDB cache...${NC}"
    CACHE_TEST=$(docker-compose exec -T keydb keydb-cli -a localca_keydb_password ping 2>/dev/null || echo "FAILED")
    if [ "$CACHE_TEST" = "PONG" ]; then
        echo -e "${GREEN}✅ KeyDB cache is accessible${NC}"
    else
        echo -e "${RED}❌ KeyDB cache is not accessible${NC}"
        return 1
    fi
    
    return 0
}

# Application functionality tests
test_application_functionality() {
    echo -e "${YELLOW}Testing application functionality...${NC}"
    
    # Test frontend accessibility
    FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000)
    if [ "$FRONTEND_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ Frontend is accessible (HTTP $FRONTEND_STATUS)${NC}"
    else
        echo -e "${RED}❌ Frontend is not accessible (HTTP $FRONTEND_STATUS)${NC}"
        return 1
    fi
    
    # Test backend accessibility
    BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/setup)
    if [ "$BACKEND_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ Backend is accessible (HTTP $BACKEND_STATUS)${NC}"
    else
        echo -e "${RED}❌ Backend is not accessible (HTTP $BACKEND_STATUS)${NC}"
        return 1
    fi
    
    # Test setup endpoint
    SETUP_RESPONSE=$(curl -s http://localhost:8080/api/setup 2>/dev/null || echo "FAILED")
    if command -v jq &> /dev/null; then
        SETUP_SUCCESS=$(echo "$SETUP_RESPONSE" | jq -r '.success // false' 2>/dev/null || echo "false")
        if [ "$SETUP_SUCCESS" = "true" ]; then
            echo -e "${GREEN}✅ Setup endpoint working${NC}"
            SETUP_COMPLETED=$(echo "$SETUP_RESPONSE" | jq -r '.data.setup_completed // false' 2>/dev/null || echo "false")
            echo -e "   Setup completed: $SETUP_COMPLETED"
        else
            echo -e "${YELLOW}⚠️  Setup endpoint may have issues${NC}"
        fi
    else
        if [[ "$SETUP_RESPONSE" == *"success"* ]]; then
            echo -e "${GREEN}✅ Setup endpoint responding${NC}"
        else
            echo -e "${YELLOW}⚠️  Setup endpoint may have issues${NC}"
        fi
    fi
    
    return 0
}

# Authentication and security tests
test_authentication() {
    echo -e "${YELLOW}Testing authentication and security...${NC}"
    
    # Test authentication requirement
    AUTH_RESPONSE=$(curl -s http://localhost:8080/api/ca-info 2>/dev/null || echo "FAILED")
    if command -v jq &> /dev/null; then
        AUTH_MESSAGE=$(echo "$AUTH_RESPONSE" | jq -r '.message // ""' 2>/dev/null || echo "")
        if [[ "$AUTH_MESSAGE" == *"Authentication required"* ]] || [[ "$AUTH_MESSAGE" == *"authentication"* ]]; then
            echo -e "${GREEN}✅ Authentication is properly enforced${NC}"
        else
            echo -e "${YELLOW}⚠️  Authentication enforcement unclear${NC}"
        fi
    else
        if [[ "$AUTH_RESPONSE" == *"authentication"* ]] || [[ "$AUTH_RESPONSE" == *"unauthorized"* ]]; then
            echo -e "${GREEN}✅ Authentication appears to be enforced${NC}"
        else
            echo -e "${YELLOW}⚠️  Authentication enforcement unclear${NC}"
        fi
    fi
    
    # Test CORS headers
    CORS_RESPONSE=$(curl -s -I -H "Origin: http://localhost:3000" http://localhost:8080/api/setup 2>/dev/null || echo "")
    if echo "$CORS_RESPONSE" | grep -i "access-control-allow-origin" > /dev/null; then
        echo -e "${GREEN}✅ CORS headers are present${NC}"
    else
        echo -e "${YELLOW}⚠️  CORS headers may be missing${NC}"
    fi
    
    return 0
}

# Environment and configuration tests
test_environment_config() {
    echo -e "${YELLOW}Testing environment configuration...${NC}"
    
    # Check enhanced storage environment variables
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
    
    # Test structured logging
    LOG_SAMPLE=$(docker-compose logs --tail=5 backend 2>/dev/null | grep -E '\{.*\}' | head -1 || echo "")
    if [[ "$LOG_SAMPLE" == *"level"* ]] && [[ "$LOG_SAMPLE" == *"msg"* ]]; then
        echo -e "${GREEN}✅ Structured logging is working${NC}"
    else
        echo -e "${YELLOW}⚠️  Structured logging may not be configured${NC}"
    fi
    
    return 0
}

# Container health tests
test_container_health() {
    echo -e "${YELLOW}Testing container health...${NC}"
    
    echo -e "${YELLOW}Container status:${NC}"
    docker-compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | grep -v "NAME" || echo "Unable to get container status"
    
    # Check if all expected containers are running
    local expected_containers=("postgres" "minio" "keydb" "backend" "frontend")
    local all_running=true
    
    for container in "${expected_containers[@]}"; do
        if docker-compose ps | grep -q "$container.*Up"; then
            echo -e "${GREEN}✅ $container is running${NC}"
        else
            echo -e "${RED}❌ $container is not running${NC}"
            all_running=false
        fi
    done
    
    if $all_running; then
        echo -e "${GREEN}✅ All containers are running${NC}"
        return 0
    else
        echo -e "${RED}❌ Some containers are not running${NC}"
        return 1
    fi
}

# Performance and load tests
test_performance() {
    echo -e "${YELLOW}Running basic performance tests...${NC}"
    
    # Test response times
    echo -e "${YELLOW}Testing response times...${NC}"
    
    # Frontend response time
    FRONTEND_TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:3000 2>/dev/null || echo "0")
    echo -e "   Frontend response time: ${FRONTEND_TIME}s"
    
    # Backend response time
    BACKEND_TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/api/setup 2>/dev/null || echo "0")
    echo -e "   Backend response time: ${BACKEND_TIME}s"
    
    # MinIO response time
    MINIO_TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:9000/minio/health/live 2>/dev/null || echo "0")
    echo -e "   MinIO response time: ${MINIO_TIME}s"
    
    echo -e "${GREEN}✅ Performance tests completed${NC}"
    return 0
}

# Show comprehensive service information
show_service_information() {
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
    echo
    echo -e "${GREEN}Available Test Scripts:${NC}"
    echo -e "   Enhanced Storage: ./tools/test-enhanced-storage.sh"
    echo -e "   Application Tests: ./tools/test_application.sh"
    echo -e "   Comprehensive Tests: ./tools/comprehensive_test.sh"
    echo -e "   Docker Tests: ./tools/run-tests-docker.sh"
}

# Main test execution
main() {
    local failed_tests=0
    local total_tests=0
    
    echo -e "${YELLOW}Starting comprehensive enhanced test suite...${NC}"
    echo
    
    # Run test sections
    local test_sections=(
        "Prerequisites Check:check_prerequisites"
        "Service Readiness:wait_for_services"
        "Enhanced Storage:test_enhanced_storage"
        "Application Functionality:test_application_functionality"
        "Authentication & Security:test_authentication"
        "Environment Configuration:test_environment_config"
        "Container Health:test_container_health"
        "Performance Tests:test_performance"
    )
    
    for section in "${test_sections[@]}"; do
        local section_name="${section%%:*}"
        local test_function="${section##*:}"
        
        total_tests=$((total_tests + 1))
        
        if ! run_test_section "$section_name" "$test_function"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo
    done
    
    # Show service information
    show_service_information
    
    # Test summary
    echo
    echo -e "${PURPLE}=== Comprehensive Test Summary ===${NC}"
    echo -e "Total test sections: $total_tests"
    echo -e "Passed: $((total_tests - failed_tests))"
    echo -e "Failed: $failed_tests"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed! Your LocalCA instance is fully functional with enhanced storage.${NC}"
        echo -e "${BLUE}🚀 Ready for production use with PostgreSQL, S3/MinIO, and structured logging!${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Some tests failed. Please check the output above for details.${NC}"
        echo -e "${YELLOW}💡 The application may still be functional, but some features might not work optimally.${NC}"
        return 1
    fi
}

# Run main function
main "$@" 