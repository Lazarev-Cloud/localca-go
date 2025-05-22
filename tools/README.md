# LocalCA Development Tools

This directory contains various development, testing, and deployment scripts for the LocalCA project with enhanced storage capabilities.

## Enhanced Storage Features

LocalCA now includes enterprise-grade storage and logging:

### 🗄️ PostgreSQL Database
- Certificate metadata storage with GORM ORM
- Comprehensive audit logging with IP/user agent tracking
- CA information management with key hashing
- Email settings persistence
- Serial number mapping for fast lookups

### 📦 S3/MinIO Object Storage
- Certificate file backup and distributed storage
- AWS S3 and MinIO compatibility
- Automatic bucket management and content-type detection
- Presigned URL generation for secure access

### ⚡ KeyDB Cache
- Redis-compatible high-performance caching
- Configurable TTL and connection pooling
- Performance optimization for certificate operations

### 📊 Structured Logging
- JSON format logs with logrus
- Separate audit log streams
- Configurable output destinations (stdout/stderr/file)
- Operation-specific log helpers

## Scripts

### Development Scripts
- `run-dev.sh` / `run-dev.bat` - Start development environment
- `run-docker.sh` / `run-docker.bat` - **ENHANCED** - Run with Docker including PostgreSQL, MinIO, KeyDB
- `deployment-status.sh` - **NEW** - Comprehensive deployment status and health overview

### Testing Scripts

#### Core Testing
- `run-tests.sh` / `run-tests.bat` - Run Go tests with coverage
- `run-tests-docker.sh` / `run-tests-docker.bat` - Run comprehensive Docker test suite
- `test_application.sh` - Basic application functionality tests

#### Enhanced Storage Testing
- `test-enhanced-storage.sh` - **NEW** - Test PostgreSQL, S3/MinIO, KeyDB, and structured logging
- `comprehensive-enhanced-test.sh` - **NEW** - Complete test suite with all enhanced features
- `comprehensive_test.sh` - Original comprehensive application testing

#### Validation Scripts
- `simple-validation.sh` - Basic validation tests
- `validate-workflows.sh` - Validate GitHub Actions workflows
- `fix-workflows.sh` - Fix workflow configurations

### Security Scripts
- `run-security-scan.sh` / `run-security-scan.bat` - Security vulnerability scanning
- `syft-install.sh` - Install Syft for SBOM generation

## Quick Start Guide

### 1. Deploy with Enhanced Storage
```bash
# Start all services (PostgreSQL, MinIO, KeyDB, LocalCA)
./tools/run-docker.sh
```

### 2. Test Enhanced Storage Features
```bash
# Test database, S3, cache, and logging
./tools/test-enhanced-storage.sh
```

### 3. Run Comprehensive Tests
```bash
# Full system validation including performance tests
./tools/comprehensive-enhanced-test.sh
```

## Service Access

After deployment with `run-docker.sh`:

| Service | URL | Purpose |
|---------|-----|---------|
| **Frontend** | http://localhost:3000 | Web interface |
| **Backend API** | http://localhost:8080 | REST API |
| **MinIO Console** | http://localhost:9001 | Object storage management |
| **PostgreSQL** | localhost:5432 | Database access |
| **KeyDB Cache** | localhost:6379 | Cache access |

## Default Credentials

| Service | Username | Password |
|---------|----------|----------|
| **PostgreSQL** | localca | localca_postgres_password |
| **MinIO** | localca | localca_minio_password |
| **KeyDB** | - | localca_keydb_password |

## Environment Variables

The enhanced storage features are controlled by environment variables (set in docker-compose.yml):

```bash
# Database Configuration
DATABASE_ENABLED=true
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=localca
DATABASE_USER=localca
DATABASE_PASSWORD=localca_postgres_password

# S3/MinIO Configuration
S3_ENABLED=true
S3_ENDPOINT=minio:9000
S3_ACCESS_KEY=localca
S3_SECRET_KEY=localca_minio_password
S3_BUCKET_NAME=localca-certificates

# Cache Configuration
CACHE_ENABLED=true
KEYDB_HOST=keydb
KEYDB_PORT=6379
KEYDB_PASSWORD=localca_keydb_password

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

## Testing Strategy

### 1. Basic Functionality
```bash
./tools/test_application.sh
```

### 2. Enhanced Storage Features
```bash
./tools/test-enhanced-storage.sh
```
Tests:
- PostgreSQL connectivity and table creation
- MinIO bucket operations and health
- KeyDB cache read/write operations
- Structured logging validation
- Environment variable configuration

### 3. Comprehensive System Validation
```bash
./tools/comprehensive-enhanced-test.sh
```
Tests:
- All enhanced storage features
- Application functionality
- Authentication and security
- Container health
- Performance metrics
- CORS and API endpoints

### 4. Docker-based Testing
```bash
./tools/run-tests-docker.sh
```

## Troubleshooting

📖 **For comprehensive troubleshooting, see [docs/TROUBLESHOOTING.md](../docs/TROUBLESHOOTING.md)**

### Common Issues

1. **Services not starting:** Check Docker Compose logs
   ```bash
   docker-compose logs
   ```

2. **Database connection issues:** Verify PostgreSQL health
   ```bash
   docker-compose exec postgres pg_isready -U localca
   ```

3. **MinIO access issues:** Check MinIO health endpoint
   ```bash
   curl http://localhost:9000/minio/health/live
   ```

4. **Cache connection issues:** Test KeyDB connectivity
   ```bash
   docker-compose exec keydb keydb-cli -a localca_keydb_password ping
   ```

### Health Checks

All services include health checks that can be monitored:
```bash
docker-compose ps
```

### Log Analysis

View structured logs from the backend:
```bash
docker-compose logs backend | jq .
```

## Usage

Make sure scripts are executable:
```bash
chmod +x tools/*.sh
```

Run from project root:
```bash
# Deploy with enhanced storage
./tools/run-docker.sh

# Test enhanced storage features
./tools/test-enhanced-storage.sh

# Run comprehensive tests
./tools/comprehensive-enhanced-test.sh
```

## Migration from File-Only Storage

If upgrading from a file-only LocalCA installation:

1. Backup existing data: `cp -r ./data ./data_backup`
2. Run enhanced deployment: `./tools/run-docker.sh`
3. Test migration: `./tools/test-enhanced-storage.sh`

The application automatically:
- Creates database tables via GORM migrations
- Creates S3 buckets if they don't exist
- Migrates existing certificate metadata
- Maintains file storage as primary/fallback