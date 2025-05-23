# LocalCA with PostgreSQL and S3/MinIO Setup Guide

This guide explains how to set up LocalCA with PostgreSQL database and S3/MinIO storage for enhanced logging, audit trails, and certificate storage.

## Overview

LocalCA now supports three storage backends:
1. **File Storage** (default, always available as fallback)
2. **PostgreSQL Database** (for structured data, audit logs, certificate metadata)
3. **S3/MinIO Storage** (for certificate files, backups, and distributed storage)

## Quick Start with Docker Compose

The easiest way to get started is using the provided `docker-compose.yml` which includes all services:

```bash
# Clone the repository
git clone <repository-url>
cd localca-go

# Start all services (PostgreSQL, MinIO, KeyDB, LocalCA)
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f backend
```

### Services Included

- **PostgreSQL** (port 5432): Database for structured data and audit logs
- **MinIO** (ports 9000, 9001): S3-compatible object storage for certificate files
- **KeyDB** (port 6379): Redis-compatible cache for performance
- **LocalCA Backend** (ports 8080, 8443, 8555): Main application
- **LocalCA Frontend** (port 3000): Web interface

### Default Credentials

- **PostgreSQL**: `localca` / `localca_postgres_password`
- **MinIO**: `localca` / `localca_minio_password`
- **KeyDB**: `localca_keydb_password`

## Manual Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 12+ (optional)
- MinIO or AWS S3 (optional)
- Redis/KeyDB (optional, for caching)

### Environment Variables

#### Database Configuration

```bash
# Enable PostgreSQL database
export DATABASE_ENABLED=true
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=localca
export DATABASE_USER=localca
export DATABASE_PASSWORD=your_secure_password
export DATABASE_SSL_MODE=disable  # or require/verify-full for production
```

#### S3/MinIO Configuration

```bash
# Enable S3/MinIO storage
export S3_ENABLED=true
export S3_ENDPOINT=localhost:9000  # MinIO endpoint or s3.amazonaws.com for AWS
export S3_ACCESS_KEY=your_access_key
export S3_SECRET_KEY=your_secret_key
export S3_BUCKET_NAME=localca-certificates
export S3_USE_SSL=false  # true for production/AWS S3
export S3_REGION=us-east-1
```

#### Cache Configuration (Optional)

```bash
# Enable KeyDB/Redis caching
export CACHE_ENABLED=true
export KEYDB_HOST=localhost
export KEYDB_PORT=6379
export KEYDB_PASSWORD=your_cache_password
export KEYDB_DB=0
export CACHE_TTL=3600
```

#### Logging Configuration

```bash
# Structured logging
export LOG_LEVEL=info          # debug, info, warn, error
export LOG_FORMAT=json         # json or text
export LOG_OUTPUT=stdout       # stdout, stderr, or file path
```

#### Core Application Settings

```bash
# Basic configuration
export CA_NAME="Your CA Name"
export ORGANIZATION="Your Organization"
export COUNTRY=US
export DATA_DIR=./data
export LISTEN_ADDR=:8080
export TLS_ENABLED=true
```

### Database Setup

1. **Install PostgreSQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql
   
   # Or use Docker
   docker run --name localca-postgres -e POSTGRES_PASSWORD=localca_postgres_password -p 5432:5432 -d postgres:16-alpine
   ```

2. **Create Database and User**:
   ```sql
   CREATE DATABASE localca;
   CREATE USER localca WITH ENCRYPTED PASSWORD 'localca_postgres_password';
   GRANT ALL PRIVILEGES ON DATABASE localca TO localca;
   ```

3. **Database Schema**: The application automatically creates tables on first run using GORM migrations.

### MinIO Setup

1. **Install MinIO**:
   ```bash
   # Download MinIO server
   wget https://dl.min.io/server/minio/release/linux-amd64/minio
   chmod +x minio
   
   # Or use Docker
   docker run -p 9000:9000 -p 9001:9001 --name localca-minio \
     -e "MINIO_ROOT_USER=localca" \
     -e "MINIO_ROOT_PASSWORD=localca_minio_password" \
     minio/minio server /data --console-address ":9001"
   ```

2. **Create Bucket**: The application automatically creates the bucket on first run.

3. **Access MinIO Console**: Visit `http://localhost:9001` to manage buckets and files.

## Features

### Database Features

- **Certificate Metadata**: Store certificate details, expiration dates, revocation status
- **Audit Logging**: Complete audit trail of all operations with timestamps, IP addresses, user agents
- **CA Information**: Centralized CA metadata with key hashing for verification
- **Email Settings**: Persistent email configuration storage
- **Serial Number Mapping**: Fast certificate lookup by serial number

### S3/MinIO Features

- **Certificate Storage**: Automatic backup of all certificate files (.pem, .key, .p12, .txt)
- **CA Backup**: Secure storage of CA certificates and encrypted keys
- **Presigned URLs**: Secure, time-limited access to certificate files
- **Bucket Organization**: Structured storage with prefixes (ca/, certificates/)
- **Content Type Detection**: Proper MIME types for different file formats

### Enhanced Logging

- **Structured Logging**: JSON format with consistent fields
- **Audit Events**: Separate audit log stream for compliance
- **Multiple Outputs**: Console, file, or syslog output
- **Log Levels**: Configurable verbosity (debug, info, warn, error)

## Storage Backends

### File Storage (Always Available)
- Primary storage for certificates and CA files
- Used as fallback when database/S3 are unavailable
- Local filesystem with proper permissions

### Database Storage (Optional)
- Structured data storage with ACID compliance
- Fast queries and reporting capabilities
- Automatic migrations and schema management
- Connection pooling and health checks

### S3 Storage (Optional)
- Distributed, scalable object storage
- Automatic bucket creation and management
- Compatible with AWS S3, MinIO, and other S3-compatible services
- Graceful fallback to file storage if unavailable

## Health Monitoring

The application provides health checks for all storage backends:

```bash
# Check application health
curl http://localhost:8080/health

# View storage backend status in logs
docker-compose logs backend | grep "Storage backend"
```

## Security Considerations

### Database Security
- Use strong passwords for database connections
- Enable SSL/TLS for database connections in production
- Regularly backup database with encrypted backups
- Implement database access controls and monitoring

### S3/MinIO Security
- Use IAM roles and policies for AWS S3
- Enable bucket encryption at rest
- Configure bucket policies for least privilege access
- Enable access logging and monitoring
- Use HTTPS/TLS for all S3 communications in production

### Application Security
- Store sensitive configuration in environment variables or secrets management
- Use encrypted CA private keys with strong passwords
- Implement proper file permissions (600 for private keys, 644 for certificates)
- Enable audit logging for compliance requirements

## Backup and Recovery

### Database Backup
```bash
# Create database backup
pg_dump -h localhost -U localca localca > localca_backup.sql

# Restore database
psql -h localhost -U localca localca < localca_backup.sql
```

### S3/MinIO Backup
```bash
# Using MinIO client (mc)
mc mirror localca-minio/localca-certificates ./backup/certificates/

# Using AWS CLI (for S3)
aws s3 sync s3://localca-certificates ./backup/certificates/
```

### File Storage Backup
```bash
# Backup data directory
tar -czf localca_files_backup.tar.gz ./data/
```

## Troubleshooting

### Database Connection Issues
- Check PostgreSQL service status
- Verify connection parameters and credentials
- Check firewall and network connectivity
- Review database logs for authentication errors

### S3/MinIO Connection Issues
- Verify endpoint URL and credentials
- Check bucket permissions and policies
- Test connectivity with MinIO client or AWS CLI
- Review S3 service logs

### Performance Issues
- Enable caching with KeyDB/Redis
- Monitor database query performance
- Check S3 request patterns and costs
- Review application logs for bottlenecks

### Common Error Messages

**"Database is not enabled"**: Set `DATABASE_ENABLED=true`
**"S3 storage is not enabled"**: Set `S3_ENABLED=true`
**"Failed to connect to database"**: Check database connection parameters
**"Failed to create S3 client"**: Verify S3 credentials and endpoint

## Migration from File-Only Storage

If you're upgrading from a file-only LocalCA installation:

1. **Backup existing data**:
   ```bash
   cp -r ./data ./data_backup
   ```

2. **Enable new storage backends** by setting environment variables

3. **Start the application** - it will automatically:
   - Create database tables
   - Create S3 buckets
   - Migrate existing certificate metadata to database
   - Upload existing certificate files to S3

4. **Verify migration** by checking logs and storage backends

## Environment Variable Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_ENABLED` | `false` | Enable PostgreSQL database |
| `DATABASE_HOST` | `localhost` | PostgreSQL host |
| `DATABASE_PORT` | `5432` | PostgreSQL port |
| `DATABASE_NAME` | `localca` | Database name |
| `DATABASE_USER` | `localca` | Database username |
| `DATABASE_PASSWORD` | - | Database password (required) |
| `DATABASE_SSL_MODE` | `disable` | SSL mode (disable/require/verify-full) |
| `S3_ENABLED` | `false` | Enable S3/MinIO storage |
| `S3_ENDPOINT` | `localhost:9000` | S3 endpoint URL |
| `S3_ACCESS_KEY` | - | S3 access key (required) |
| `S3_SECRET_KEY` | - | S3 secret key (required) |
| `S3_BUCKET_NAME` | `localca-certificates` | S3 bucket name |
| `S3_USE_SSL` | `false` | Use HTTPS for S3 connections |
| `S3_REGION` | `us-east-1` | S3 region |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log format (json/text) |
| `LOG_OUTPUT` | `stdout` | Log output (stdout/stderr/file) |

## Support

For issues and questions:
1. Check the application logs for error messages
2. Verify environment variable configuration
3. Test individual storage backend connectivity
4. Review this documentation for common solutions
5. Open an issue on the project repository 