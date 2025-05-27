@echo off
setlocal enabledelayedexpansion

echo ========================================
echo LocalCA Integration Fix Script
echo ========================================
echo.

echo [1/6] Ensuring required files exist...

REM Ensure data directory exists
if not exist ".\data" (
    echo Creating data directory...
    mkdir .\data
)

REM Ensure cakey.txt exists
if not exist ".\data\cakey.txt" (
    echo Creating cakey.txt with secure password...
    echo LocalCA_SecurePassword_2024! > .\data\cakey.txt
)
echo ✓ Required files exist

echo.
echo [2/6] Stopping existing containers...
docker-compose down >nul 2>&1
echo ✓ Containers stopped

echo.
echo [3/6] Cleaning Docker cache...
docker system prune -f >nul 2>&1
echo ✓ Docker cache cleaned

echo.
echo [4/6] Building containers with no cache...
docker-compose build --no-cache
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to build containers
    exit /b 1
)
echo ✓ Containers built successfully

echo.
echo [5/6] Starting services...
docker-compose up -d
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to start containers
    exit /b 1
)
echo ✓ Services started

echo.
echo [6/6] Waiting for services to initialize...
echo Waiting 45 seconds for all services to be ready...
timeout /t 45 /nobreak >nul

echo.
echo ========================================
echo Testing Integration
echo ========================================

REM Test backend
echo Testing backend API...
curl -s -o nul -w "Backend Status: %%{http_code}" http://localhost:8080/api/auth/status
echo.

REM Test frontend
echo Testing frontend...
curl -s -o nul -w "Frontend Status: %%{http_code}" http://localhost:3000
echo.

REM Test proxy
echo Testing API proxy...
curl -s -o nul -w "Proxy Status: %%{http_code}" http://localhost:3000/api/proxy/api/auth/status
echo.

REM Test CA download endpoint specifically
echo Testing CA download endpoint...
curl -s -o nul -w "CA Download Status: %%{http_code}" http://localhost:3000/api/proxy/api/download/ca
echo.

echo.
echo ========================================
echo Setup Instructions
echo ========================================
echo.
echo 1. Open http://localhost:3000 in your browser
echo 2. If you see a setup page, get the setup token:
echo    docker-compose logs backend ^| findstr "Setup Token"
echo 3. Complete the setup process
echo 4. After setup, the CA download should work
echo.
echo Common URLs:
echo - Frontend: http://localhost:3000
echo - Backend API: http://localhost:8080
echo - Direct CA download: http://localhost:8080/api/download/ca
echo - Proxy CA download: http://localhost:3000/api/proxy/api/download/ca
echo.
echo To view logs: docker-compose logs -f
echo To stop: docker-compose down
echo.

endlocal 