# LocalCA Docker Setup Guide

This guide will help you get LocalCA up and running using Docker with all components (Go backend, Next.js frontend, PostgreSQL, MinIO, and KeyDB).

## Prerequisites

- Docker Desktop installed and running
- Docker Compose v2.0+ installed
- At least 4GB of available RAM
- Ports 3000, 8080, 8443, 8555, 5432, 6379, 9000, 9001 available

## Quick Start

### Option 1: Using the Startup Script (Recommended)

**Windows:**
```cmd
start-localca.bat
```

**Linux/macOS:**
```bash
./start-localca.sh
```

### Option 2: Manual Docker Compose

1. **Create data directory and CA key file:**
```bash
mkdir -p data
echo "LocalCA_SecurePassword_2024!" > data/cakey.txt
```

2. **Start all services:**
```bash
docker-compose up -d
```

3. **Check service status:**
```bash
docker-compose ps
```

## Services Overview

LocalCA runs the following services:

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 3000 | Next.js web interface |
| Backend | 8080 | Go API server (HTTP) |
| Backend | 8443 | Go API server (HTTPS) |
| ACME Server | 8555 | ACME protocol endpoint |
| PostgreSQL | 5432 | Database storage |
| MinIO | 9000 | S3-compatible object storage |
| MinIO Console | 9001 | MinIO web interface |
| KeyDB | 6379 | Redis-compatible cache |

## Initial Setup

1. **Wait for services to start** (about 30-60 seconds)

2. **Get the setup token:**
```bash
docker-compose logs backend | grep "Setup token"
```

3. **Open the web interface:**
   - Navigate to http://localhost:3000
   - You'll be redirected to the setup page

4. **Complete the setup:**
   - Enter the setup token from step 2
   - Create an admin username and password
   - Configure your Certificate Authority details

## Service Credentials

### Database (PostgreSQL)
- **Host:** localhost:5432
- **Database:** localca
- **Username:** localca
- **Password:** localca_postgres_password

### Object Storage (MinIO)
- **Console:** http://localhost:9001
- **API:** http://localhost:9000
- **Access Key:** localca
- **Secret Key:** localca_minio_password

### Cache (KeyDB)
- **Host:** localhost:6379
- **Password:** localca_keydb_password

## Configuration

The Docker Compose setup includes comprehensive configuration for all services. Key environment variables are set in the `docker-compose.yml` file:

### Backend Configuration
- Enhanced storage with PostgreSQL, MinIO, and KeyDB
- CORS configured for frontend communication
- TLS enabled for secure connections
- Structured JSON logging
- Comprehensive audit logging

### Frontend Configuration
- Next.js 15 with App Router
- API proxy routes for backend communication
- Production-optimized build
- Responsive design with dark/light themes

## Useful Commands

### Service Management
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Restart a specific service
docker-compose restart backend

# View service status
docker-compose ps

# Scale a service (if needed)
docker-compose up -d --scale backend=2
```

### Logs and Debugging
```bash
# View all logs
docker-compose logs

# View logs for a specific service
docker-compose logs backend
docker-compose logs frontend

# Follow logs in real-time
docker-compose logs -f

# View last 50 lines of logs
docker-compose logs --tail=50
```

### Data Management
```bash
# Backup data directory
tar -czf localca-backup-$(date +%Y%m%d).tar.gz data/

# View Docker volumes
docker volume ls

# Inspect a volume
docker volume inspect localca-go_postgres-data
```

## Troubleshooting

### Common Issues

#### 1. Port Conflicts
If you get port binding errors:
```bash
# Check what's using the ports
netstat -tulpn | grep :3000
netstat -tulpn | grep :8080

# Stop conflicting services or change ports in docker-compose.yml
```

#### 2. Services Not Starting
```bash
# Check Docker daemon
docker info

# Check service logs
docker-compose logs [service-name]

# Restart Docker Desktop (Windows/macOS)
```

#### 3. Database Connection Issues
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Connect to database directly
docker-compose exec postgres psql -U localca -d localca
```

#### 4. Frontend Not Loading
```bash
# Check frontend logs
docker-compose logs frontend

# Verify backend is running
curl http://localhost:8080/api/health
```

### Health Checks

All services include health checks. Check service health:
```bash
# View service health status
docker-compose ps

# Check specific service health
docker inspect localca-backend --format='{{.State.Health.Status}}'
```

### Performance Optimization

#### For Development
- Reduce resource limits in docker-compose.yml
- Use bind mounts for faster file changes
- Enable hot reloading for frontend

#### For Production
- Use Docker secrets for sensitive data
- Configure resource limits appropriately
- Set up log rotation
- Use external databases for better performance

## Security Considerations

### Default Security Features
- CA private key password protection
- TLS encryption for API communications
- Session-based authentication with secure cookies
- CSRF protection
- Input validation and sanitization
- Comprehensive audit logging

### Production Security
1. **Change default passwords** in docker-compose.yml
2. **Use Docker secrets** for sensitive configuration
3. **Enable TLS** for all external communications
4. **Configure firewall rules** to restrict access
5. **Regular security updates** for base images
6. **Monitor logs** for security events

## Backup and Recovery

### Automated Backup
```bash
# Create backup script
cat > backup-localca.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="backups/localca_$DATE"

mkdir -p "$BACKUP_DIR"

# Backup data directory
cp -r data/ "$BACKUP_DIR/"

# Backup database
docker-compose exec -T postgres pg_dump -U localca localca > "$BACKUP_DIR/database.sql"

# Backup MinIO data
docker-compose exec -T minio mc mirror /data "$BACKUP_DIR/minio/"

echo "Backup completed: $BACKUP_DIR"
EOF

chmod +x backup-localca.sh
```

### Recovery
```bash
# Restore from backup
BACKUP_DIR="backups/localca_20241201_120000"

# Stop services
docker-compose down

# Restore data
cp -r "$BACKUP_DIR/data/" ./

# Start services
docker-compose up -d

# Restore database (if needed)
docker-compose exec -T postgres psql -U localca localca < "$BACKUP_DIR/database.sql"
```

## Development

### Local Development Setup
```bash
# Use development compose file
docker-compose -f docker-compose.dev.yml up -d

# Or run services individually
docker-compose up -d postgres minio keydb  # Infrastructure only
go run main.go                              # Backend locally
npm run dev                                 # Frontend locally
```

### Building Custom Images
```bash
# Build specific service
docker-compose build backend

# Build with no cache
docker-compose build --no-cache

# Build and start
docker-compose up -d --build
```

## Monitoring

### Service Monitoring
```bash
# Monitor resource usage
docker stats

# Monitor specific containers
docker stats localca-backend localca-frontend

# View system events
docker events
```

### Application Monitoring
- Backend health: http://localhost:8080/api/health
- Frontend health: http://localhost:3000/api/health
- Database metrics: Available through PostgreSQL logs
- Cache metrics: Available through KeyDB INFO command

## Support

For issues and questions:
1. Check the logs: `docker-compose logs`
2. Review this documentation
3. Check the main README.md
4. Open an issue on the project repository

## Next Steps

After successful setup:
1. **Create your first certificate** through the web interface
2. **Configure ACME clients** to use your LocalCA
3. **Set up automated certificate renewal**
4. **Configure monitoring and alerting**
5. **Implement backup procedures** 