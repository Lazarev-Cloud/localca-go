@echo off
echo Running security scan for LocalCA-Go...

echo Generating test coverage...
go test ./... -coverprofile=coverage.out -json > test-report.out

echo Running SonarQube scan...
if "%SONAR_TOKEN%"=="" (
    echo SONAR_TOKEN environment variable is not set. Please set it before running this script.
    exit /b 1
)

sonar-scanner.bat -Dsonar.login=%SONAR_TOKEN%

echo Security scan completed. 