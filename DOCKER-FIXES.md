# Docker Deployment Fixes

This document details the fixes made to the Docker deployment configuration for better frontend-backend communication.

## Issues Fixed

1. **CORS Configuration**:
   - Updated CORS middleware to properly use environment variables
   - Added support for multiple allowed origins
   - Added `Access-Control-Allow-Credentials: true` to support cookie-based auth

2. **Cookie Handling**:
   - Fixed cookie domain and security settings for Docker environment
   - Improved cookie forwarding between Next.js API routes and backend
   - Added environment variables for cookie configuration
   - Updated the frontend cookie handling to ensure cookies are properly set

3. **API URL Configuration**:
   - Properly configured API URL for container networking (using `http://backend:8080` inside docker network)
   - This ensures API calls from the frontend container can reach the backend container
   - For browser clients, all API requests go through Next.js API routes

4. **Connection Error Handling**:
   - Added timeout handlers to fetch calls to prevent hanging requests
   - Improved error handling for connection failures
   - Added retry capabilities for transient network errors

5. **Authentication Redirects**:
   - Enhanced the setup and login redirect handling
   - Fixed redirect loops by adding state management
   - Added better error handling for auth failures

6. **Deployment Simplification**:
   - Added deployment scripts for macOS/Linux (`run-docker.sh`) and Windows (`run-docker.bat`)
   - Scripts handle data directory creation and CA key generation
   - Added helpful startup messages

## Using the Fixed Deployment

1. **Run the deployment script**:
   ```bash
   # macOS/Linux
   ./run-docker.sh
   
   # Windows
   run-docker.bat
   ```

2. **Complete the setup**:
   - Navigate to http://localhost:3000/setup
   - Get the setup token from the logs: `docker-compose logs backend | grep "Setup token"`
   - Complete the admin account creation

3. **Login and use the application**:
   - Access the UI at http://localhost:3000
   - API directly accessible at http://localhost:8080

## Environment Variables

The following environment variables can be configured in `docker-compose.yml`:

- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins
- `CORS_ALLOWED_METHODS`: Comma-separated list of allowed HTTP methods
- `CORS_ALLOWED_HEADERS`: Comma-separated list of allowed HTTP headers
- `COOKIE_DOMAIN`: Domain for cookies (blank for current domain)
- `COOKIE_SECURE`: Whether cookies should be secure-only (true/false)
- `NEXT_PUBLIC_API_URL`: Backend API URL for the frontend

## Container Communication Architecture

The solution uses a layered approach to ensure proper communication:

1. **Browser → Frontend Container**:
   - Browser makes requests to Next.js server at http://localhost:3000
   - Next.js serves the frontend application

2. **Frontend Container → Backend Container**:
   - Frontend API routes forward requests to backend using Docker network hostname
   - Communication happens through internal Docker network (`localca-network`)
   - Uses `http://backend:8080` URL for internal container-to-container communication

3. **Next.js API Routes**: 
   - Act as proxies between browser and backend
   - Handle cookie forwarding and error transformation
   - Implement timeouts and connection error handling

## Troubleshooting

If you encounter issues:

1. **Check logs**: `docker-compose logs`
2. **Access API directly**: Try accessing http://localhost:8080/api/ca-info
3. **Check cookies**: Ensure cookies are being set correctly in browser dev tools
4. **Restart containers**: `docker-compose restart`
5. **Network issues**: Make sure containers can communicate by running `docker-compose exec frontend ping backend` 