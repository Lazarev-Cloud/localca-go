# Integration Testing with LocalCA

This document describes how to run integration tests that work with the actual Go backend instead of mocks.

## Overview

The integration tests are designed to:
- Start a real Go backend server
- Run Next.js tests against the actual API
- Test the complete authentication flow
- Verify API proxy functionality
- Clean up test environment automatically

## Quick Start

### Linux/macOS
```bash
chmod +x run-integration-tests.sh
./run-integration-tests.sh
```

### Windows
```cmd
run-integration-tests.bat
```

### Using npm directly
```bash
npm run test:integration
```

## Test Structure

### Configuration Files
- `jest.config.integration.js` - Jest configuration for integration tests
- `jest.setup.integration.js` - Test environment setup without mocks
- `test-utils/global-setup.js` - Starts Go backend before tests
- `test-utils/global-teardown.js` - Stops backend and cleans up after tests
- `test-utils/global.d.ts` - TypeScript declarations for test utilities

### Test Files
- `__tests__/integration/auth.integration.test.tsx` - Authentication flow tests
- `__tests__/integration/api.integration.test.ts` - API endpoint tests

## How It Works

### 1. Global Setup
The global setup process:
1. Creates a test data directory
2. Sets up CA key password file
3. Configures environment variables for testing
4. Builds the Go backend (`localca-test`)
5. Starts the backend server
6. Waits for the backend to be ready

### 2. Test Execution
Each test:
1. Waits for backend to be ready
2. Resets backend state (clears sessions)
3. Runs the actual test against the live backend
4. Uses real HTTP requests instead of mocks

### 3. Global Teardown
The cleanup process:
1. Stops the Go backend gracefully
2. Removes test files and certificates
3. Cleans up the test binary

## Environment Variables

The integration tests use these environment variables:

```bash
# Test Backend Configuration
DATA_DIR=./data
CA_KEY_FILE=./data/cakey.txt
CA_NAME=LocalCA Test
ORGANIZATION=LocalCA Test Org
COUNTRY=US
TLS_ENABLED=false
EMAIL_NOTIFY=false
DEBUG=true
GIN_MODE=debug
CORS_ALLOWED_ORIGINS=http://localhost:*
ALLOW_LOCALHOST=true
COOKIE_SECURE=false

# Disable external services for testing
DATABASE_ENABLED=false
S3_ENABLED=false
CACHE_ENABLED=false

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
NODE_ENV=test
```

## Test Utilities

### Global Functions
Available in all integration tests:

```typescript
// Wait for backend to be ready
await global.waitForBackend(url?: string, timeout?: number)

// Reset backend state (clear sessions)
await global.resetBackendState()

// Test configuration
global.testConfig = {
  backendUrl: 'http://localhost:8080',
  timeout: 30000,
  retryAttempts: 3,
  retryDelay: 1000
}
```

## Test Categories

### Authentication Tests
- Initial setup flow
- Login with correct/incorrect credentials
- Field validation
- Authentication state management
- Session handling

### API Tests
- Direct backend API calls
- Next.js API proxy functionality
- Error handling
- CORS configuration
- Complete authentication flow

## Running Specific Tests

```bash
# Run only authentication tests
npm run test:integration -- --testNamePattern="Authentication"

# Run only API tests
npm run test:integration -- --testNamePattern="API"

# Run with verbose output
npm run test:integration -- --verbose

# Run in watch mode
npm run test:integration:watch
```

## Debugging

### Backend Logs
The backend output is displayed during test execution with `[Backend]` prefix.

### Test Timeouts
Integration tests have longer timeouts (60 seconds) to account for:
- Backend startup time
- Real HTTP requests
- Certificate generation

### Common Issues

1. **Backend fails to start**
   - Check if port 8080 is available
   - Ensure Go is installed and in PATH
   - Check data directory permissions

2. **Tests timeout**
   - Backend might be slow to start
   - Check backend logs for errors
   - Increase timeout in test configuration

3. **Authentication failures**
   - Backend state might not be reset properly
   - Check if setup was completed in previous runs
   - Clear data directory manually if needed

## Differences from Unit Tests

| Aspect | Unit Tests | Integration Tests |
|--------|------------|-------------------|
| Backend | Mocked with `jest-fetch-mock` | Real Go server |
| Speed | Fast (~seconds) | Slower (~minutes) |
| Isolation | Complete | Shared backend state |
| Setup | Minimal | Full environment |
| Reliability | High | Dependent on system |

## Best Practices

1. **Test Independence**: Each test should reset backend state
2. **Error Handling**: Tests should handle backend startup failures gracefully
3. **Cleanup**: Always clean up test data and processes
4. **Timeouts**: Use appropriate timeouts for real network calls
5. **Logging**: Include helpful debug information for failures

## Continuous Integration

For CI environments, consider:
- Pre-building the Go backend
- Using Docker for consistent environment
- Parallel test execution limitations
- Resource cleanup on failure

## Troubleshooting

### Port Conflicts
If port 8080 is in use:
```bash
# Find process using port 8080
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Kill the process or change test port
```

### Permission Issues
```bash
# Make scripts executable
chmod +x run-integration-tests.sh
chmod +x test-utils/*.js
```

### Clean Reset
```bash
# Remove all test data
rm -rf data/
rm -f localca-test localca-test.exe
``` 