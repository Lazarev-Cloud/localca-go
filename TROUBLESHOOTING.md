# LocalCA Troubleshooting Guide

## Common Issues and Solutions

### 404 Error: `GET http://localhost:3000/api/proxy/api/download/ca 404 (Not Found)`

This error indicates that the frontend cannot reach the backend API through the proxy. Here are the steps to resolve it:

#### Step 1: Verify Docker Setup

1. **Check if Docker is running:**
   ```cmd
   docker info
   ```

2. **Ensure all containers are running:**
   ```cmd
   docker-compose ps
   ```
   You should see all services (backend, frontend, postgres, minio, keydb) as "Up".

3. **Check container logs:**
   ```cmd
   docker-compose logs backend
   docker-compose logs frontend
   ```

#### Step 2: Verify Required Files

1. **Check if `cakey.txt` exists:**
   ```cmd
   dir data\cakey.txt
   ```
   If missing, create it:
   ```cmd
   echo LocalCA_SecurePassword_2024! > data\cakey.txt
   ```

#### Step 3: Test API Endpoints

1. **Test backend directly:**
   ```cmd
   curl http://localhost:8080/api/auth/status
   ```
   Expected: 401 (Unauthorized) or 200 (OK)

2. **Test frontend:**
   ```cmd
   curl http://localhost:3000
   ```
   Expected: 200 (OK)

3. **Test proxy:**
   ```cmd
   curl http://localhost:3000/api/proxy/api/auth/status
   ```
   Expected: 401 (Unauthorized) or 200 (OK)

4. **Test CA download specifically:**
   ```cmd
   curl -I http://localhost:3000/api/proxy/api/download/ca
   ```
   Expected: 401 (setup required) or 200 (OK after setup)

#### Step 4: Fix Common Issues

**Issue: Backend not starting**
- Check if port 8080 is already in use
- Verify `cakey.txt` exists in the data directory
- Check backend logs: `docker-compose logs backend`

**Issue: Frontend proxy not working**
- Verify `NEXT_PUBLIC_API_URL=http://backend:8080` in docker-compose.yml
- Check frontend logs: `docker-compose logs frontend`
- Restart containers: `docker-compose restart`

**Issue: Network connectivity**
- Ensure all containers are on the same network
- Check Docker network: `docker network ls`
- Verify container connectivity: `docker-compose exec frontend ping backend`

#### Step 5: Complete Setup Process

1. **Get setup token:**
   ```cmd
   docker-compose logs backend | findstr "Setup Token"
   ```

2. **Open browser and complete setup:**
   - Go to http://localhost:3000
   - If redirected to setup, use the token from step 1
   - Create admin credentials
   - Complete setup

3. **Test CA download after setup:**
   - Login to the application
   - Try downloading CA certificate from the dashboard

## Quick Fix Script

Run the comprehensive fix script:
```cmd
tools\fix-integration.bat
```

This script will:
1. Create required files
2. Stop existing containers
3. Clean Docker cache
4. Rebuild containers
5. Start services
6. Test all endpoints

## Environment Variables

### Backend (docker-compose.yml)
```yaml
environment:
  - NEXT_PUBLIC_API_URL=http://backend:8080  # Internal Docker network
  - CA_KEY_FILE=/app/data/cakey.txt
  - DATA_DIR=/app/data
  - CORS_ALLOWED_ORIGINS=http://localhost:*,https://localhost:*,http://frontend:3000
```

### Frontend (docker-compose.yml)
```yaml
environment:
  - NEXT_PUBLIC_API_URL=http://backend:8080  # Internal Docker network
  - USE_PROXY_ROUTES=true
```

## API Endpoint Mapping

| Frontend URL | Proxy Route | Backend URL |
|--------------|-------------|-------------|
| `/api/proxy/api/download/ca` | `app/api/proxy/[...path]/route.ts` | `http://backend:8080/api/download/ca` |
| `/api/proxy/api/auth/status` | `app/api/proxy/[...path]/route.ts` | `http://backend:8080/api/auth/status` |
| `/api/proxy/api/certificates` | `app/api/proxy/[...path]/route.ts` | `http://backend:8080/api/certificates` |

## Certificate Workflow

1. **Initial Setup:**
   - Backend creates CA certificate on first run
   - Setup token is generated
   - User completes setup via frontend

2. **CA Download:**
   - Frontend: User clicks "Download CA"
   - Request: `GET /api/proxy/api/download/ca`
   - Proxy: Forwards to `http://backend:8080/api/download/ca`
   - Backend: Returns CA certificate file

3. **Certificate Management:**
   - Create, renew, revoke certificates through frontend
   - All operations go through the proxy to backend

## Debugging Commands

```cmd
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Check container status
docker-compose ps

# Restart specific service
docker-compose restart backend
docker-compose restart frontend

# Rebuild and restart everything
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Test network connectivity
docker-compose exec frontend ping backend
docker-compose exec backend ping frontend

# Check environment variables
docker-compose exec backend env | grep API
docker-compose exec frontend env | grep API
```

## Support

If issues persist:
1. Check the container logs for specific error messages
2. Verify all environment variables are set correctly
3. Ensure no other services are using ports 3000, 8080, 8443, or 8555
4. Try running the fix script: `tools\fix-integration.bat` 