@echo off
setlocal enabledelayedexpansion

echo ========================================
echo LocalCA Docker Integration Test
echo ========================================
echo.

REM Check if Docker is running
echo [1/8] Checking Docker status...
docker info >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: Docker is not running. Please start Docker Desktop.
    exit /b 1
)
echo ✓ Docker is running

REM Check if cakey.txt exists
echo.
echo [2/8] Checking CA key file...
if not exist ".\data\cakey.txt" (
    echo Creating cakey.txt with secure password...
    echo LocalCA_SecurePassword_2024! > .\data\cakey.txt
)
echo ✓ CA key file exists

REM Stop any existing containers
echo.
echo [3/8] Stopping existing containers...
docker-compose down >nul 2>&1
echo ✓ Existing containers stopped

REM Build and start containers
echo.
echo [4/8] Building and starting containers...
docker-compose build --no-cache
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to build containers
    exit /b 1
)

docker-compose up -d
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to start containers
    exit /b 1
)
echo ✓ Containers started

REM Wait for services to be ready
echo.
echo [5/8] Waiting for services to be ready...
timeout /t 30 /nobreak >nul
echo ✓ Services should be ready

REM Test backend health
echo.
echo [6/8] Testing backend health...
curl -s -o nul -w "%%{http_code}" http://localhost:8080/api/auth/status > temp_status.txt
set /p BACKEND_STATUS=<temp_status.txt
del temp_status.txt

if "!BACKEND_STATUS!"=="200" (
    echo ✓ Backend is healthy (200)
) else if "!BACKEND_STATUS!"=="401" (
    echo ✓ Backend is healthy (401 - setup required)
) else (
    echo ✗ Backend health check failed (Status: !BACKEND_STATUS!)
    echo Showing backend logs:
    docker-compose logs --tail=20 backend
)

REM Test frontend health
echo.
echo [7/8] Testing frontend health...
curl -s -o nul -w "%%{http_code}" http://localhost:3000 > temp_status.txt
set /p FRONTEND_STATUS=<temp_status.txt
del temp_status.txt

if "!FRONTEND_STATUS!"=="200" (
    echo ✓ Frontend is healthy (200)
) else (
    echo ✗ Frontend health check failed (Status: !FRONTEND_STATUS!)
    echo Showing frontend logs:
    docker-compose logs --tail=20 frontend
)

REM Test API proxy
echo.
echo [8/8] Testing API proxy integration...
curl -s -o nul -w "%%{http_code}" http://localhost:3000/api/proxy/api/auth/status > temp_status.txt
set /p PROXY_STATUS=<temp_status.txt
del temp_status.txt

if "!PROXY_STATUS!"=="200" (
    echo ✓ API proxy is working (200)
) else if "!PROXY_STATUS!"=="401" (
    echo ✓ API proxy is working (401 - setup required)
) else (
    echo ✗ API proxy test failed (Status: !PROXY_STATUS!)
    echo This is likely the cause of the 404 error you're seeing
)

echo.
echo ========================================
echo Test Results Summary
echo ========================================
echo Backend Status: !BACKEND_STATUS!
echo Frontend Status: !FRONTEND_STATUS!
echo Proxy Status: !PROXY_STATUS!
echo.

if "!PROXY_STATUS!"=="401" (
    echo ✓ All services are working correctly!
    echo.
    echo Next steps:
    echo 1. Open http://localhost:3000 in your browser
    echo 2. Complete the initial setup
    echo 3. The CA download should work after setup
    echo.
    echo To get the setup token, run:
    echo docker-compose logs backend ^| findstr "Setup Token"
) else (
    echo ✗ There are issues with the setup
    echo.
    echo Troubleshooting:
    echo 1. Check container logs: docker-compose logs
    echo 2. Verify all containers are running: docker-compose ps
    echo 3. Check network connectivity between containers
)

echo.
echo To view logs: docker-compose logs -f
echo To stop services: docker-compose down
echo.

endlocal 