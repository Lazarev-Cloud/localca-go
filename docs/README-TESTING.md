# LocalCA Testing Guide

This guide will help you run tests and verify the build process for LocalCA using Docker.

## Prerequisites

- Docker installed and running
- Docker Compose installed

## Running Tests in Docker

### For Linux/macOS Users

1. Make the scripts executable:
   ```bash
   chmod +x run-tests-docker.sh scripts/verify-build.sh
   ```

2. Run all tests with the provided script:
   ```bash
   ./run-tests-docker.sh
   ```

### For Windows Users

1. Run all tests with the provided batch script:
   ```powershell
   .\run-tests-docker.bat
   ```

## What Gets Tested

The test suite includes:

1. **Backend Tests** - Runs all Go tests for the application
2. **Frontend Tests** - Runs Jest tests for the Next.js frontend
3. **Build Verification** - Ensures both Go and Next.js components can build successfully

## Running Individual Tests

You can run individual test components using Docker Compose directly:

```bash
# Run only backend tests
docker-compose -f docker-compose.test.yml run --rm backend-test

# Run only frontend tests
docker-compose -f docker-compose.test.yml run --rm frontend-test

# Run only build verification
docker-compose -f docker-compose.test.yml run --rm build-check
```

## Running the Complete Application

To run the full application locally:

```bash
# Start the application
docker-compose up -d

# Check the status
docker-compose ps

# View logs
docker-compose logs -f

# Stop the application
docker-compose down
```

## Troubleshooting

- **Dependencies Issue**: If you encounter any dependency issues, the testing Dockerfiles already include compatibility fixes for testing libraries.
- **Permission Issues**: On Linux/macOS, if you encounter permission issues, make sure the scripts are executable with `chmod +x`.
- **Docker Issues**: Make sure Docker and Docker Compose are installed and running correctly.

## Files Overview

- `docker-compose.test.yml` - Configuration for testing in Docker
- `Dockerfile.test-backend` - Docker configuration for Go backend tests
- `Dockerfile.test-frontend` - Docker configuration for Next.js frontend tests
- `Dockerfile.test-build` - Docker configuration for build verification
- `scripts/verify-build.sh` - Script that checks if the application builds correctly
- `run-tests-docker.sh` - Shell script for running all tests (Linux/macOS)
- `run-tests-docker.bat` - Batch script for running all tests (Windows) 