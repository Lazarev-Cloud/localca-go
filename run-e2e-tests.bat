@echo off
setlocal enabledelayedexpansion

REM LocalCA End-to-End Integration Test Runner for Windows
REM This script runs comprehensive integration tests with Docker backend

echo 🚀 Starting LocalCA End-to-End Integration Tests
echo ==================================================

REM Check prerequisites
echo [STEP] Checking prerequisites...

where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed or not in PATH
    exit /b 1
)

where docker-compose >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Docker Compose is not installed or not in PATH
    exit /b 1
)

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed or not in PATH
    exit /b 1
)

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] npm is not installed or not in PATH
    exit /b 1
)

echo [INFO] All prerequisites found

REM Display versions
echo [STEP] Environment information:
docker --version
docker-compose --version
node --version
npm --version

REM Install dependencies
echo [STEP] Installing/updating npm dependencies...
npm install
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install npm dependencies
    exit /b 1
)

REM Clean up any existing containers
echo [STEP] Cleaning up existing test containers...
docker-compose -f docker-compose.test.yml down --remove-orphans >nul 2>nul

REM Remove old test data
echo [STEP] Cleaning up old test data...
if exist test-data rmdir /s /q test-data >nul 2>nul

REM Create test data directory
echo [STEP] Setting up test environment...
mkdir test-data >nul 2>nul
echo test-ca-password-123 > test-data\cakey.txt

REM Start Docker backend
echo [STEP] Starting Docker backend for testing...
docker-compose -f docker-compose.test.yml up --build -d
if %errorlevel% neq 0 (
    echo [ERROR] Failed to start Docker backend
    exit /b 1
)

REM Wait for backend to be ready
echo [STEP] Waiting for backend to be ready...
set BACKEND_URL=http://localhost:8080
set MAX_WAIT=120
set /a WAIT_COUNT=0

:wait_loop
curl -s -f "%BACKEND_URL%/api/ca-info" >nul 2>nul
if %errorlevel% equ 0 (
    echo [INFO] ✅ Backend is ready!
    goto backend_ready
)

set /a WAIT_COUNT+=1
if %WAIT_COUNT% geq %MAX_WAIT% (
    echo [ERROR] ❌ Backend failed to start within %MAX_WAIT% seconds
    echo [STEP] Showing backend logs:
    docker-compose -f docker-compose.test.yml logs backend-test
    goto cleanup_and_exit
)

echo|set /p="."
timeout /t 3 /nobreak >nul
goto wait_loop

:backend_ready

REM Show backend status
echo [STEP] Backend status check:
curl -s "%BACKEND_URL%/api/setup" 2>nul | findstr "success" >nul
if %errorlevel% equ 0 (
    echo Setup endpoint: accessible
) else (
    echo Setup endpoint: error
)

curl -s "%BACKEND_URL%/api/ca-info" 2>nul | findstr "success" >nul
if %errorlevel% equ 0 (
    echo CA info endpoint: accessible
) else (
    echo CA info endpoint: error
)

REM Run the integration tests
echo [STEP] Running integration tests...
echo This will test:
echo   ✓ Backend API endpoints
echo   ✓ Frontend component rendering
echo   ✓ Setup and login workflows
echo   ✓ Error handling
echo   ✓ CORS configuration
echo.

REM Run tests with detailed output
npm run test:integration -- --verbose --detectOpenHandles
if %errorlevel% equ 0 (
    echo [INFO] ✅ All integration tests passed!
    
    REM Run additional manual verification
    echo [STEP] Running manual verification checks...
    
    REM Test setup endpoint
    echo [STEP] Testing setup endpoint...
    curl -s "%BACKEND_URL%/api/setup"
    echo.
    
    REM Test CA info endpoint
    echo [STEP] Testing CA info endpoint...
    curl -s "%BACKEND_URL%/api/ca-info"
    echo.
    
    REM Test login endpoint
    echo [STEP] Testing login endpoint...
    curl -s -X POST "%BACKEND_URL%/api/login" -H "Content-Type: application/json" -d "{\"username\":\"admin\",\"password\":\"wrongpass\"}"
    echo.
    
    echo [INFO] 🎉 All tests completed successfully!
    
) else (
    echo [ERROR] ❌ Integration tests failed!
    
    echo [STEP] Showing backend logs for debugging:
    docker-compose -f docker-compose.test.yml logs backend-test
    
    goto cleanup_and_exit
)

goto cleanup_and_success

:cleanup_and_exit
echo [STEP] Cleaning up test environment...
docker-compose -f docker-compose.test.yml down --remove-orphans >nul 2>nul
if exist test-data rmdir /s /q test-data >nul 2>nul
echo [INFO] Cleanup completed
exit /b 1

:cleanup_and_success
echo [STEP] Cleaning up test environment...
docker-compose -f docker-compose.test.yml down --remove-orphans >nul 2>nul
if exist test-data rmdir /s /q test-data >nul 2>nul
echo [INFO] Cleanup completed
exit /b 0 