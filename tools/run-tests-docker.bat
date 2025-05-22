@echo off
setlocal

echo ===== LocalCA Docker Tests =====

REM Ensure scripts directory exists
if not exist scripts mkdir scripts

REM Clean up any existing test containers
echo Cleaning up previous test containers...
docker-compose -f docker-compose.test.yml down -v 2>nul

REM Build containers first to avoid long waits during testing
echo Building test containers (this may take a minute)...
docker-compose -f docker-compose.test.yml build
if %ERRORLEVEL% neq 0 (
    echo [31m❌ Building test containers failed[0m
    exit /b 1
)
echo [32m✅ Test containers built successfully[0m

REM Run backend tests with verbose output
echo Running backend tests (this may take a minute)...
docker-compose -f docker-compose.test.yml run --rm backend-test
if %ERRORLEVEL% neq 0 (
    echo [31m❌ Backend tests failed[0m
    exit /b 1
)
echo [32m✅ Backend tests passed[0m

REM Run frontend tests with verbose output
echo Running frontend tests...
docker-compose -f docker-compose.test.yml run --rm frontend-test
if %ERRORLEVEL% neq 0 (
    echo [31m❌ Frontend tests failed[0m
    exit /b 1
)
echo [32m✅ Frontend tests passed[0m

REM Verify build with more verbose output
echo Verifying build process...
docker-compose -f docker-compose.test.yml run --rm build-check
if %ERRORLEVEL% neq 0 (
    echo [31m❌ Build verification failed[0m
    exit /b 1
)
echo [32m✅ Build verification passed[0m

echo [32m===== All tests passed! =====[0m

REM Clean up test containers and volumes
echo Cleaning up test containers and volumes...
docker-compose -f docker-compose.test.yml down -v

exit /b 0 