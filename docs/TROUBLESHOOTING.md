# LocalCA Troubleshooting Guide

This guide covers common issues and solutions for LocalCA deployment and usage.

## Table of Contents

- [Authentication Issues](#authentication-issues)
- [Setup and Initial Configuration](#setup-and-initial-configuration)
- [Enhanced Storage Issues](#enhanced-storage-issues)
- [Container and Docker Issues](#container-and-docker-issues)
- [Certificate Management Issues](#certificate-management-issues)
- [Network and Connectivity Issues](#network-and-connectivity-issues)
- [Performance Issues](#performance-issues)
- [Logging and Debugging](#logging-and-debugging)

## Authentication Issues

### 🔐 Login Credentials Not Working (401 Unauthorized)

**Problem**: Getting "Invalid username or password" error when trying to log in with expected credentials.

**Symptoms**:
- 401 Unauthorized error in browser
- "Invalid username or password" message
- Unable to access the web interface

**Solution**: Reset the authentication setup

1. **Stop the backend container**:
   ```bash
   docker-compose stop backend
   ```

2. **Remove the auth configuration file**:
   ```bash
   docker-compose exec backend rm /app/data/auth.json
   # Or if container is stopped:
   docker run --rm -v localca-go_data:/data alpine rm /data/auth.json
   ```

3. **Restart the backend**:
   ```bash
   docker-compose start backend
   # Wait a few seconds for startup
   sleep 3
   ```

4. **Get the new setup token**:
   ```bash
   curl -s http://localhost:8080/api/setup | jq .
   ```

5. **Complete setup with desired credentials**:
   ```bash
   curl -X POST 'http://localhost:8080/api/setup' \
     -H 'Content-Type: application/json' \
     -d '{
       "username": "admin",
       "password": "test123",
       "confirm_password": "test123",
       "setup_token": "YOUR_SETUP_TOKEN_HERE"
     }'
   ```

6. **Verify login works**:
   ```bash
   curl -X POST 'http://localhost:8080/api/login' \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     -d 'username=admin&password=test123'
   ```

**Default Working Credentials**:
- Username: `admin`
- Password: `test123`

### 🔄 Session Expired or Invalid

**Problem**: Getting logged out frequently or session not persisting.

**Solution**:
1. Clear browser cookies for localhost:3000
2. Ensure containers are running: `docker-compose ps`
3. Check session directory: `docker-compose exec backend ls -la /app/data/sessions/`
4. Restart backend if sessions directory is corrupted:
   ```bash
   docker-compose restart backend
   ```

## Setup and Initial Configuration

### 🚀 Setup Token Not Found

**Problem**: Cannot find the setup token needed for initial configuration.

**Solutions**:

1. **Check backend logs**:
   ```bash
   docker-compose logs backend | grep -i "setup\|token"
   ```

2. **Get token via API**:
   ```bash
   curl -s http://localhost:8080/api/setup | jq .data.setup_token
   ```

3. **If setup is already completed but you need to reset**:
   ```bash
   # Remove auth file and restart
   docker-compose exec backend rm /app/data/auth.json
   docker-compose restart backend
   ```

### 🔧 Setup Already Completed Error

**Problem**: Trying to run setup but getting "Setup is already completed" error.

**Solution**: Reset the setup process (see Authentication Issues section above).

## Enhanced Storage Issues

### 🗄️ PostgreSQL Database Connection Issues

**Problem**: Database connection failures or table creation errors.

**Diagnosis**:
```bash
# Check PostgreSQL health
docker-compose exec postgres pg_isready -U localca -d localca

# Check database logs
docker-compose logs postgres

# Test connection manually
docker-compose exec postgres psql -U localca -d localca -c "SELECT version();"
```

**Solutions**:

1. **Database not ready**: Wait for health check to pass
   ```bash
   # Wait for healthy status
   docker-compose ps | grep postgres
   ```

2. **Connection refused**: Restart PostgreSQL
   ```bash
   docker-compose restart postgres
   ```

3. **Tables not created**: Check GORM migrations
   ```bash
   # Check if tables exist
   docker-compose exec postgres psql -U localca -d localca -c "\dt"
   ```

4. **Permission issues**: Recreate the database volume
   ```bash
   docker-compose down
   docker volume rm localca-go_postgres-data
   docker-compose up -d
   ```

### 📦 MinIO S3 Storage Issues

**Problem**: MinIO not accessible or bucket operations failing.

**Diagnosis**:
```bash
# Check MinIO health
curl -f http://localhost:9000/minio/health/live

# Check MinIO logs
docker-compose logs minio

# Test bucket access
docker-compose exec minio mc alias set local http://localhost:9000 localca localca_minio_password
docker-compose exec minio mc ls local/
```

**Solutions**:

1. **MinIO not starting**: Check port conflicts
   ```bash
   # Check if ports 9000/9001 are in use
   lsof -i :9000
   lsof -i :9001
   ```

2. **Bucket not found**: Recreate bucket
   ```bash
   docker-compose exec minio mc mb local/localca-certificates
   ```

3. **Access denied**: Reset MinIO credentials
   ```bash
   docker-compose restart minio
   ```

### ⚡ KeyDB Cache Issues

**Problem**: Cache connection failures or performance issues.

**Diagnosis**:
```bash
# Test KeyDB connection
docker-compose exec keydb keydb-cli -a localca_keydb_password ping

# Check memory usage
docker-compose exec keydb keydb-cli -a localca_keydb_password info memory

# Check KeyDB logs
docker-compose logs keydb
```

**Solutions**:

1. **Connection refused**: Restart KeyDB
   ```bash
   docker-compose restart keydb
   ```

2. **Memory issues**: Clear cache
   ```bash
   docker-compose exec keydb keydb-cli -a localca_keydb_password flushall
   ```

3. **Authentication failed**: Check password in docker-compose.yml

## Container and Docker Issues

### 🐳 Containers Not Starting

**Problem**: Docker containers failing to start or crashing.

**Diagnosis**:
```bash
# Check container status
docker-compose ps

# Check logs for errors
docker-compose logs

# Check resource usage
docker stats
```

**Solutions**:

1. **Port conflicts**: Change ports in docker-compose.yml
2. **Resource constraints**: Increase Docker memory/CPU limits
3. **Volume issues**: Remove and recreate volumes
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

### 🔄 Build Failures

**Problem**: Docker build process failing.

**Solutions**:

1. **Clear Docker cache**:
   ```bash
   docker system prune -a
   docker-compose build --no-cache
   ```

2. **Check Dockerfile syntax**: Ensure all files exist
3. **Network issues**: Check internet connection for package downloads

## Certificate Management Issues

### 📜 Certificate Creation Failures

**Problem**: Unable to create certificates through the web interface.

**Diagnosis**:
```bash
# Check CA certificate exists
docker-compose exec backend ls -la /app/data/ca.pem

# Check CA private key
docker-compose exec backend ls -la /app/data/cakey.txt

# Check backend logs for errors
docker-compose logs backend | grep -i error
```

**Solutions**:

1. **CA not initialized**: Restart backend to trigger CA creation
2. **Permission issues**: Check file permissions in data directory
3. **Invalid domain names**: Ensure proper FQDN format

### 🔒 Certificate Download Issues

**Problem**: Cannot download certificates or getting 404 errors.

**Solutions**:

1. **Check certificate exists**:
   ```bash
   docker-compose exec backend ls -la /app/data/[certificate-name]/
   ```

2. **Verify file permissions**: Ensure files are readable
3. **Clear browser cache**: Force refresh the download page

## Network and Connectivity Issues

### 🌐 Frontend Not Accessible

**Problem**: Cannot access http://localhost:3000

**Solutions**:

1. **Check frontend container**:
   ```bash
   docker-compose ps frontend
   docker-compose logs frontend
   ```

2. **Port conflicts**: Change frontend port in docker-compose.yml
3. **Firewall issues**: Check local firewall settings

### 🔌 Backend API Not Responding

**Problem**: Cannot access http://localhost:8080

**Solutions**:

1. **Check backend health**:
   ```bash
   curl -f http://localhost:8080/api/setup
   ```

2. **Check backend logs**:
   ```bash
   docker-compose logs backend
   ```

3. **Restart backend**:
   ```bash
   docker-compose restart backend
   ```

### 🔗 CORS Issues

**Problem**: Cross-origin request errors in browser console.

**Solutions**:

1. **Check CORS headers**: Verify in browser developer tools
2. **Use proxy routes**: Access API through frontend proxy at `/api/*`
3. **Check environment variables**: Ensure proper API URL configuration

## Performance Issues

### 🐌 Slow Response Times

**Problem**: Web interface or API responding slowly.

**Diagnosis**:
```bash
# Check response times
./tools/deployment-status.sh

# Check container resource usage
docker stats

# Check database performance
docker-compose exec postgres psql -U localca -d localca -c "SELECT COUNT(*) FROM certificates;"
```

**Solutions**:

1. **Increase container resources**: Modify docker-compose.yml
2. **Clear cache**: Restart KeyDB
3. **Database optimization**: Check for large tables
4. **Restart services**:
   ```bash
   docker-compose restart
   ```

### 💾 High Memory Usage

**Problem**: Containers consuming too much memory.

**Solutions**:

1. **Check memory usage**:
   ```bash
   docker stats --no-stream
   ```

2. **Restart memory-heavy containers**:
   ```bash
   docker-compose restart minio backend
   ```

3. **Clear logs**:
   ```bash
   docker system prune
   ```

## Logging and Debugging

### 📊 Enable Debug Logging

**Problem**: Need more detailed logs for troubleshooting.

**Solution**: Set debug environment variables in docker-compose.yml:

```yaml
environment:
  - LOG_LEVEL=debug
  - GIN_MODE=debug
```

Then restart:
```bash
docker-compose restart backend
```

### 📝 View Structured Logs

**Problem**: Need to analyze JSON logs.

**Solutions**:

1. **View formatted logs**:
   ```bash
   docker-compose logs backend | jq .
   ```

2. **Filter logs by level**:
   ```bash
   docker-compose logs backend | jq 'select(.level=="error")'
   ```

3. **Search logs**:
   ```bash
   docker-compose logs backend | grep -i "certificate\|error"
   ```

### 🔍 Audit Trail Analysis

**Problem**: Need to track certificate operations.

**Solutions**:

1. **Check audit logs in database**:
   ```bash
   docker-compose exec postgres psql -U localca -d localca -c "SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;"
   ```

2. **Export audit logs**:
   ```bash
   docker-compose exec postgres psql -U localca -d localca -c "COPY audit_logs TO STDOUT WITH CSV HEADER;" > audit_export.csv
   ```

## Quick Diagnostic Commands

### 🔧 Health Check Script

Run this comprehensive health check:

```bash
# Use the built-in deployment status script
./tools/deployment-status.sh

# Or run individual checks
./tools/test-enhanced-storage.sh
./tools/comprehensive-enhanced-test.sh
```

### 🚨 Emergency Reset

If everything is broken, perform a complete reset:

```bash
# Stop all services
docker-compose down

# Remove all data (WARNING: This deletes all certificates!)
docker volume rm localca-go_postgres-data localca-go_minio-data
rm -rf ./data/*

# Restart fresh
./tools/run-docker.sh

# Complete setup again
# Visit http://localhost:3000/setup
```

## Getting Help

### 📞 Support Resources

1. **Check logs first**: `docker-compose logs`
2. **Run diagnostics**: `./tools/deployment-status.sh`
3. **Review configuration**: Check docker-compose.yml and environment variables
4. **Test connectivity**: Use curl commands provided in this guide

### 🐛 Reporting Issues

When reporting issues, include:

1. **System information**: OS, Docker version
2. **Error messages**: Full error text and stack traces
3. **Configuration**: docker-compose.yml (remove sensitive data)
4. **Logs**: Relevant container logs
5. **Steps to reproduce**: Exact commands that cause the issue

### 📚 Additional Resources

- **Setup Guide**: `docs/SETUP_DATABASE_S3.md`
- **Tools Documentation**: `tools/README.md`
- **API Documentation**: Check `/api/setup` endpoint for current status
- **Test Scripts**: Use scripts in `tools/` directory for validation

---

**Last Updated**: May 2025  
**Version**: Enhanced Storage Release 