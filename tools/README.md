# Build Configuration

This directory contains Docker build configurations and test setups.

## Files

### Test Dockerfiles
- `Dockerfile.test-backend` - Go backend testing container
- `Dockerfile.test-frontend` - Frontend testing container  
- `Dockerfile.test-build` - Build verification container

### Test Configuration
- `docker-compose.test.yml` - Docker Compose configuration for testing

## Usage

### Running Tests

From the project root, use the test scripts:

```bash
# Run all Docker tests
./tools/run-tests-docker.sh

# Or run specific test services
docker-compose -f build/docker-compose.test.yml run --rm backend-test
docker-compose -f build/docker-compose.test.yml run --rm frontend-test
docker-compose -f build/docker-compose.test.yml run --rm build-check
```

### Test Services

- **backend-test**: Runs Go tests with coverage
- **frontend-test**: Runs Jest tests for frontend components
- **build-check**: Verifies that both backend and frontend build successfully

### Cleanup

```bash
# Clean up test containers and volumes
docker-compose -f build/docker-compose.test.yml down -v
```