@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    LocalCA Docker Deployment Script
echo ========================================
echo.

REM Check if Docker is running
docker info >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Docker is not running or not installed.
    echo Please start Docker Desktop and try again.
    pause
    exit /b 1
)

REM Check if docker-compose is available
where docker-compose >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] docker-compose is not installed.
    echo Please install Docker Compose and try again.
    pause
    exit /b 1
)

echo [INFO] Docker is running and docker-compose is available.
echo.

REM Create data directory if it doesn't exist
if not exist "data" (
    echo [INFO] Creating data directory...
    mkdir data
)

REM Create cakey.txt if it doesn't exist
if not exist "data\cakey.txt" (
    echo [INFO] Creating CA key password file...
    echo LocalCA_SecurePassword_2024! > data\cakey.txt
)

echo [INFO] Stopping any existing containers...
docker-compose down >nul 2>&1

echo [INFO] Building Docker images...
docker-compose build

echo [INFO] Starting LocalCA services...
docker-compose up -d

echo.
echo [SUCCESS] LocalCA is starting up!
echo.
echo Services:
echo - Frontend UI:     http://localhost:3000
echo - Backend API:     http://localhost:8080
echo - ACME Server:     http://localhost:8555
echo - PostgreSQL:      localhost:5432
echo - MinIO Console:   http://localhost:9001
echo - KeyDB:           localhost:6379
echo.
echo Credentials:
echo - Database: localca / localca_postgres_password
echo - MinIO:    localca / localca_minio_password
echo - KeyDB:    localca_keydb_password
echo.
echo [INFO] Waiting for services to start...
timeout /t 10 /nobreak >nul

echo [INFO] Checking service status...
docker-compose ps

echo.
echo [INFO] Getting setup token from backend logs...
docker-compose logs backend | findstr "Setup token" 2>nul
if %ERRORLEVEL% neq 0 (
    echo [INFO] Setup token not found yet. Services may still be starting.
    echo Check logs with: docker-compose logs backend
)

echo.
echo [INFO] To complete setup:
echo 1. Visit http://localhost:3000/setup
echo 2. Use the setup token from the logs above
echo 3. Create your admin account
echo.
echo [INFO] To stop services: docker-compose down
echo [INFO] To view logs: docker-compose logs [service-name]
echo.
echo Press any key to show live logs...
pause >nul

docker-compose logs -f 