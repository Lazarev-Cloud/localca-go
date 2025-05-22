#!/bin/bash
set -e

echo "===== LocalCA Build Verification ====="
echo "Starting verification process..."

# Create directory if it doesn't exist
mkdir -p scripts

# Update package.json for compatibility
echo "Updating package.json for compatibility..."
jq '.devDependencies."@testing-library/react" = "^14.0.0"' package.json > tmp.json && mv tmp.json package.json
jq '.dependencies.react = "^18.2.0" | .dependencies."react-dom" = "^18.2.0"' package.json > tmp.json && mv tmp.json package.json
jq '.dependencies."date-fns" = "^3.6.0"' package.json > tmp.json && mv tmp.json package.json

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download

# Build Go application
echo "Building Go application..."
CGO_ENABLED=0 go build -v -o localca-go

# Check Go build result
if [ ! -f "localca-go" ]; then
    echo "❌ Go build failed - executable not found"
    exit 1
else
    echo "✅ Go build successful"
    ls -la localca-go
fi

# Install npm dependencies with less verbose output
echo "Installing npm dependencies..."
# Suppress deprecation warnings
npm config set loglevel error
npm install --legacy-peer-deps --no-fund --no-audit

# Build frontend
echo "Building frontend..."
# Suppress deprecation warnings
NODE_OPTIONS=--no-warnings npm run build

# Check frontend build result
if [ ! -d ".next" ]; then
    echo "❌ Next.js build failed - .next directory not found"
    exit 1
else
    echo "✅ Next.js build successful"
    du -sh .next
fi

echo "===== Build verification completed successfully ====="

# You could add more verification steps here
# For example, run basic smoke tests or check for specific files

exit 0 