@echo off
setlocal enabledelayedexpansion

REM LocalCA Integration Test Runner for Windows
REM This script runs the Next.js integration tests with the actual Go backend

echo 🚀 Starting LocalCA Integration Tests
echo ======================================

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

REM Check if Node.js is installed
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed or not in PATH
    exit /b 1
)

REM Check if npm is installed
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] npm is not installed or not in PATH
    exit /b 1
)

echo [INFO] Checking Go version...
go version

echo [INFO] Checking Node.js version...
node --version

echo [INFO] Installing/updating npm dependencies...
npm install
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install npm dependencies
    exit /b 1
)

echo [INFO] Running integration tests...
echo This will:
echo   1. Build and start the Go backend
echo   2. Run Next.js integration tests
echo   3. Clean up test environment
echo.

REM Run the integration tests
npm run test:integration
if %errorlevel% equ 0 (
    echo [INFO] ✅ All integration tests passed!
    exit /b 0
) else (
    echo [ERROR] ❌ Integration tests failed!
    exit /b 1
) 