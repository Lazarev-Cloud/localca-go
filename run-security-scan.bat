@echo off
SETLOCAL

echo Running security scans...

REM Create output directory
if not exist "security-reports" mkdir security-reports

REM Run Go security checks
echo Running gosec...
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -fmt=json -out=security-reports\gosec-results.json .\...

REM Run npm audit for frontend
echo Running npm audit...
call npm audit --json > security-reports\npm-audit.json 2>nul

REM Run Go test coverage
echo Running test coverage...
go test -coverprofile=coverage.out .\...

REM Run SonarQube scan if SONAR_TOKEN is available
if not "%SONAR_TOKEN%"=="" (
  echo Running SonarQube scan...
  
  REM Download sonar-scanner if not already installed
  if not exist "sonar-scanner" (
    echo Downloading sonar-scanner...
    powershell -Command "Invoke-WebRequest -Uri https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.8.0.2856-windows.zip -OutFile sonar-scanner.zip"
    powershell -Command "Expand-Archive -Path sonar-scanner.zip -DestinationPath ."
    ren sonar-scanner-4.8.0.2856-windows sonar-scanner
    del sonar-scanner.zip
  )
  
  REM Run sonar-scanner
  sonar-scanner\bin\sonar-scanner.bat ^
    -Dsonar.projectKey=localca-go ^
    -Dsonar.sources=. ^
    -Dsonar.host.url=https://sonarcloud.io ^
    -Dsonar.login=%SONAR_TOKEN% ^
    -Dsonar.go.coverage.reportPaths=coverage.out ^
    -Dsonar.exclusions=**/*_test.go,**/vendor/**,**/testdata/**
) else (
  echo SONAR_TOKEN not set, skipping SonarQube scan
)

echo Security scans completed. Reports saved to .\security-reports\

ENDLOCAL 