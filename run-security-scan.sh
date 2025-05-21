#!/bin/bash

# Security scan script for LocalCA

echo "Running security scans..."

# Create output directory
mkdir -p ./security-reports

# Run Go security checks
echo "Running gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -fmt=json -out=./security-reports/gosec-results.json ./...

# Run npm audit for frontend
echo "Running npm audit..."
npm audit --json > ./security-reports/npm-audit.json || true

# Run Go test coverage
echo "Running test coverage..."
go test -coverprofile=coverage.out ./...

# Run SonarQube scan if SONAR_TOKEN is available
if [ -n "$SONAR_TOKEN" ]; then
  echo "Running SonarQube scan..."
  
  # Download sonar-scanner if not already installed
  if [ ! -d "./sonar-scanner" ]; then
    echo "Downloading sonar-scanner..."
    wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.8.0.2856-linux.zip
    unzip sonar-scanner-cli-4.8.0.2856-linux.zip
    mv sonar-scanner-4.8.0.2856-linux sonar-scanner
    rm sonar-scanner-cli-4.8.0.2856-linux.zip
  fi
  
  # Run sonar-scanner
  ./sonar-scanner/bin/sonar-scanner \
    -Dsonar.projectKey=localca-go \
    -Dsonar.sources=. \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.login=$SONAR_TOKEN \
    -Dsonar.go.coverage.reportPaths=coverage.out \
    -Dsonar.exclusions=**/*_test.go,**/vendor/**,**/testdata/**
else
  echo "SONAR_TOKEN not set, skipping SonarQube scan"
fi

echo "Security scans completed. Reports saved to ./security-reports/" 