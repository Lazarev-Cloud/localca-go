#!/bin/bash

# LocalCA Deployment Status Script
# Provides comprehensive overview of the enhanced LocalCA deployment

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${PURPLE}=== LocalCA Enhanced Deployment Status ===${NC}"
echo -e "${CYAN}Complete overview of your LocalCA instance with enhanced storage${NC}"
echo

# Function to check service status
check_service_status() {
    local service_name=$1
    local url=$2
    local expected_code=${3:-200}
    
    local status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$status_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ $service_name (HTTP $status_code)${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name (HTTP $status_code)${NC}"
        return 1
    fi
}

# Function to check database connectivity
check_database() {
    local db_status=$(docker-compose exec -T postgres pg_isready -U localca -d localca 2>/dev/null | grep "accepting connections" || echo "FAILED")
    
    if [[ "$db_status" == *"accepting connections"* ]]; then
        echo -e "${GREEN}✅ PostgreSQL Database${NC}"
        
        # Check table count
        local table_count=$(docker-compose exec -T postgres psql -U localca -d localca -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' \n' || echo "0")
        echo -e "   Tables: $table_count"
        
        return 0
    else
        echo -e "${RED}❌ PostgreSQL Database${NC}"
        return 1
    fi
}

# Function to check cache
check_cache() {
    local cache_status=$(docker-compose exec -T keydb keydb-cli -a localca_keydb_password ping 2>/dev/null || echo "FAILED")
    
    if [ "$cache_status" = "PONG" ]; then
        echo -e "${GREEN}✅ KeyDB Cache${NC}"
        
        # Check memory usage
        local memory_info=$(docker-compose exec -T keydb keydb-cli -a localca_keydb_password info memory 2>/dev/null | grep "used_memory_human" | cut -d: -f2 | tr -d '\r' || echo "unknown")
        echo -e "   Memory used: $memory_info"
        
        return 0
    else
        echo -e "${RED}❌ KeyDB Cache${NC}"
        return 1
    fi
}

# Function to check MinIO and bucket
check_minio() {
    local minio_status=$(curl -s -f http://localhost:9000/minio/health/live 2>/dev/null && echo "OK" || echo "FAILED")
    
    if [ "$minio_status" = "OK" ]; then
        echo -e "${GREEN}✅ MinIO S3 Storage${NC}"
        
        # Check bucket
        local bucket_check=$(docker-compose exec -T minio sh -c "mc alias set local http://localhost:9000 localca localca_minio_password >/dev/null 2>&1 && mc ls local/localca-certificates >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_FOUND'" 2>/dev/null || echo "ERROR")
        
        if [ "$bucket_check" = "EXISTS" ]; then
            echo -e "   Bucket: localca-certificates ✅"
        else
            echo -e "   Bucket: localca-certificates ⚠️"
        fi
        
        return 0
    else
        echo -e "${RED}❌ MinIO S3 Storage${NC}"
        return 1
    fi
}

# Function to show container status
show_container_status() {
    echo -e "${BLUE}=== Container Status ===${NC}"
    
    local containers=("postgres" "minio" "keydb" "backend" "frontend")
    local all_running=true
    
    for container in "${containers[@]}"; do
        if docker-compose ps | grep -q "$container.*Up"; then
            local status=$(docker-compose ps --format "{{.Status}}" | grep -E "Up|healthy" | head -1)
            echo -e "${GREEN}✅ localca-$container${NC} - $status"
        else
            echo -e "${RED}❌ localca-$container${NC} - Not running"
            all_running=false
        fi
    done
    
    if $all_running; then
        echo -e "${GREEN}All containers are running properly${NC}"
    else
        echo -e "${YELLOW}Some containers may have issues${NC}"
    fi
}

# Function to show service URLs
show_service_urls() {
    echo -e "${BLUE}=== Service Access URLs ===${NC}"
    echo -e "${CYAN}Web Interfaces:${NC}"
    echo -e "   🌐 LocalCA Frontend:  ${YELLOW}http://localhost:3000${NC}"
    echo -e "   🔧 MinIO Console:     ${YELLOW}http://localhost:9001${NC}"
    echo
    echo -e "${CYAN}API Endpoints:${NC}"
    echo -e "   🔌 Backend API:       ${YELLOW}http://localhost:8080${NC}"
    echo -e "   🔒 HTTPS API:         ${YELLOW}https://localhost:8443${NC}"
    echo -e "   📜 ACME Server:       ${YELLOW}http://localhost:8555${NC}"
    echo
    echo -e "${CYAN}Direct Service Access:${NC}"
    echo -e "   🗄️  PostgreSQL:       ${YELLOW}localhost:5432${NC}"
    echo -e "   📦 MinIO S3:          ${YELLOW}localhost:9000${NC}"
    echo -e "   ⚡ KeyDB Cache:       ${YELLOW}localhost:6379${NC}"
}

# Function to show credentials
show_credentials() {
    echo -e "${BLUE}=== Service Credentials ===${NC}"
    echo -e "${CYAN}Database (PostgreSQL):${NC}"
    echo -e "   Username: ${YELLOW}localca${NC}"
    echo -e "   Password: ${YELLOW}localca_postgres_password${NC}"
    echo -e "   Database: ${YELLOW}localca${NC}"
    echo
    echo -e "${CYAN}Object Storage (MinIO):${NC}"
    echo -e "   Access Key: ${YELLOW}localca${NC}"
    echo -e "   Secret Key: ${YELLOW}localca_minio_password${NC}"
    echo -e "   Bucket: ${YELLOW}localca-certificates${NC}"
    echo
    echo -e "${CYAN}Cache (KeyDB):${NC}"
    echo -e "   Password: ${YELLOW}localca_keydb_password${NC}"
}

# Function to show environment status
show_environment_status() {
    echo -e "${BLUE}=== Enhanced Storage Configuration ===${NC}"
    
    # Check environment variables in backend container
    local db_enabled=$(docker-compose exec -T backend printenv DATABASE_ENABLED 2>/dev/null || echo "not_set")
    local s3_enabled=$(docker-compose exec -T backend printenv S3_ENABLED 2>/dev/null || echo "not_set")
    local cache_enabled=$(docker-compose exec -T backend printenv CACHE_ENABLED 2>/dev/null || echo "not_set")
    local log_format=$(docker-compose exec -T backend printenv LOG_FORMAT 2>/dev/null || echo "not_set")
    local log_level=$(docker-compose exec -T backend printenv LOG_LEVEL 2>/dev/null || echo "not_set")
    
    echo -e "   Database Storage: ${db_enabled} $([ "$db_enabled" = "true" ] && echo "✅" || echo "❌")"
    echo -e "   S3 Object Storage: ${s3_enabled} $([ "$s3_enabled" = "true" ] && echo "✅" || echo "❌")"
    echo -e "   Cache Layer: ${cache_enabled} $([ "$cache_enabled" = "true" ] && echo "✅" || echo "❌")"
    echo -e "   Log Format: ${log_format} $([ "$log_format" = "json" ] && echo "✅" || echo "⚠️")"
    echo -e "   Log Level: ${log_level}"
}

# Function to show available commands
show_available_commands() {
    echo -e "${BLUE}=== Available Commands ===${NC}"
    echo -e "${CYAN}Testing:${NC}"
    echo -e "   ./tools/test-enhanced-storage.sh     - Test enhanced storage features"
    echo -e "   ./tools/comprehensive-enhanced-test.sh - Complete system validation"
    echo -e "   ./tools/test_application.sh          - Basic application tests"
    echo
    echo -e "${CYAN}Management:${NC}"
    echo -e "   docker-compose logs backend          - View backend logs"
    echo -e "   docker-compose logs postgres         - View database logs"
    echo -e "   docker-compose logs minio            - View MinIO logs"
    echo -e "   docker-compose restart backend       - Restart backend service"
    echo -e "   docker-compose down                  - Stop all services"
    echo
    echo -e "${CYAN}Monitoring:${NC}"
    echo -e "   docker-compose ps                    - Container status"
    echo -e "   ./tools/deployment-status.sh         - This status overview"
}

# Function to show performance metrics
show_performance_metrics() {
    echo -e "${BLUE}=== Performance Metrics ===${NC}"
    
    # Test response times
    local frontend_time=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:3000 2>/dev/null || echo "0")
    local backend_time=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/api/setup 2>/dev/null || echo "0")
    local minio_time=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:9000/minio/health/live 2>/dev/null || echo "0")
    
    echo -e "   Frontend Response Time: ${frontend_time}s"
    echo -e "   Backend Response Time:  ${backend_time}s"
    echo -e "   MinIO Response Time:    ${minio_time}s"
    
    # Container resource usage
    echo -e "${CYAN}Container Resource Usage:${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep localca || echo "   Unable to get resource stats"
}

# Main execution
main() {
    # Check if docker-compose is running
    if ! docker-compose ps | grep -q "Up"; then
        echo -e "${RED}❌ LocalCA is not running${NC}"
        echo -e "${YELLOW}Start with: ./tools/run-docker.sh${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}=== Service Health Check ===${NC}"
    check_service_status "Frontend" "http://localhost:3000"
    check_service_status "Backend API" "http://localhost:8080/api/setup"
    check_service_status "MinIO Console" "http://localhost:9001"
    check_database
    check_cache
    check_minio
    echo
    
    show_container_status
    echo
    
    show_environment_status
    echo
    
    show_service_urls
    echo
    
    show_credentials
    echo
    
    show_performance_metrics
    echo
    
    show_available_commands
    
    echo
    echo -e "${GREEN}=== Deployment Summary ===${NC}"
    echo -e "${GREEN}🎉 LocalCA is running with enhanced storage features!${NC}"
    echo -e "${CYAN}📊 PostgreSQL Database: Structured data and audit logs${NC}"
    echo -e "${CYAN}📦 MinIO S3 Storage: Certificate files and backups${NC}"
    echo -e "${CYAN}⚡ KeyDB Cache: Performance optimization${NC}"
    echo -e "${CYAN}📝 Structured Logging: JSON format with audit trails${NC}"
    echo
    echo -e "${YELLOW}🚀 Ready for certificate management and production use!${NC}"
}

# Run main function
main "$@" 