@echo off
setlocal enabledelayedexpansion

echo Starting LocalCA Docker Deployment

REM Check if Docker is installed
where docker-compose >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: docker-compose is not installed.
    echo Please install docker-compose before running this script.
    exit /b 1
)

REM Check if data directory exists, if not create it
if not exist ".\data" (
    echo Creating data directory...
    mkdir .\data
)

REM Check if cakey.txt exists, if not create it
if not exist ".\data\cakey.txt" (
    echo Creating cakey.txt with random password...
    REM Generate a random password using PowerShell
    powershell -Command "$pwd = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 16 | ForEach-Object {[char]$_}); $pwd | Out-File -FilePath .\cakey.txt -Encoding ascii"
    REM Make sure it's moved to the right location
    copy .\cakey.txt .\data\cakey.txt >nul
)

REM Build and start the containers
echo Building and starting Docker containers...
docker-compose down 2>nul
docker-compose build
docker-compose up -d

echo Docker containers are up and running!
echo - Frontend UI: http://localhost:3000
echo - Backend API: http://localhost:8080
echo.
echo Important Notes:
echo 1. On first run, you'll need to complete the setup at http://localhost:3000/setup
echo 2. The initial setup token can be found in the logs:
echo    docker-compose logs backend | findstr "Setup token"
echo.
echo To stop the services, run:
echo docker-compose down
echo.

REM Show logs after startup
echo Showing startup logs:
docker-compose logs --tail=20

echo.
echo Done!

endlocal 