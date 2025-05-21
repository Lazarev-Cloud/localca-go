#!/bin/bash
echo "Running security scan for LocalCA-Go..."

echo "Generating test coverage..."
go test ./... -coverprofile=coverage.out -json > test-report.out

echo "Running SonarQube scan..."
if [ -z "$SONAR_TOKEN" ]; then
    echo "SONAR_TOKEN environment variable is not set. Please set it before running this script."
    exit 1
fi

sonar-scanner -Dsonar.login=$SONAR_TOKEN

echo "Security scan completed." 