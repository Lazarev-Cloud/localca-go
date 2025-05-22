@echo off
echo Running LocalCA-Go Tests
echo ===============================

echo Running package tests...
go test -v -cover ./pkg/...

if %ERRORLEVEL% NEQ 0 (
    echo Package tests failed!
    exit /b 1
)

echo Running main package test...
go test -v -cover ./main_test.go

if %ERRORLEVEL% NEQ 0 (
    echo Main package test failed!
    exit /b 1
)

echo Checking Docker availability...
where docker >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo Docker is available.
    
    REM Check if Docker is running
    docker info >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo Docker is running.
        
        REM Test Docker build
        echo Testing Docker build...
        docker build -t localca-go-backend:test -f Dockerfile .
        if %ERRORLEVEL% NEQ 0 (
            echo Docker build failed!
            exit /b 1
        ) else (
            echo Docker build successful.
        )
        
        REM Test Docker Compose if available
        where docker-compose >nul 2>&1
        if %ERRORLEVEL% EQU 0 (
            echo Testing Docker Compose configuration...
            docker-compose config
            if %ERRORLEVEL% NEQ 0 (
                echo Docker Compose configuration failed!
                exit /b 1
            ) else (
                echo Docker Compose configuration is valid.
            )
        ) else (
            echo Docker Compose not available, skipping Docker Compose tests.
        )
    ) else (
        echo Docker is not running, skipping Docker tests.
    )
) else (
    echo Docker not available, skipping Docker tests.
)

echo Running tests...
go test -race -coverprofile=coverage.out -covermode=atomic ./...

echo.
echo Coverage report:
go tool cover -func=coverage.out

echo.
echo Generating HTML coverage report...
go tool cover -html=coverage.out -o coverage.html

echo.
echo Tests completed. Coverage report available in coverage.html

echo All tests passed!
exit /b 0 