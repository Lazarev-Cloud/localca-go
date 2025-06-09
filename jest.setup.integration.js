// Import Jest DOM utilities
import '@testing-library/jest-dom';

// Polyfill fetch for Node.js environment
if (!global.fetch) {
  const fetch = require('node-fetch');
  global.fetch = fetch;
  global.Headers = fetch.Headers;
  global.Request = fetch.Request;
  global.Response = fetch.Response;
}

// Mock the next/router for navigation testing
jest.mock('next/router', () => ({
  useRouter() {
    return {
      route: '/',
      pathname: '',
      query: {},
      asPath: '',
      push: jest.fn(),
      replace: jest.fn(),
      reload: jest.fn(),
      back: jest.fn(),
      prefetch: jest.fn(),
      beforePopState: jest.fn(),
      events: {
        on: jest.fn(),
        off: jest.fn(),
        emit: jest.fn(),
      },
      isFallback: false,
    };
  },
}));

// Mock next/navigation for App Router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    refresh: jest.fn(),
  })),
  usePathname: jest.fn(() => '/'),
  useSearchParams: jest.fn(() => new URLSearchParams()),
}));

// Mock next/image
jest.mock('next/image', () => ({
  __esModule: true,
  default: (props) => {
    // eslint-disable-next-line jsx-a11y/alt-text
    return <img {...props} />;
  },
}));

// Configure test environment variables
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080';
process.env.NODE_ENV = 'test';

// Global test configuration
global.testConfig = {
  backendUrl: 'http://localhost:8080',
  timeout: 60000, // Increased for Docker startup
  retryAttempts: 3,
  retryDelay: 1000,
};

// Mock the config module to use test backend URL
jest.mock('@/lib/config', () => ({
  __esModule: true,
  default: {
    apiUrl: 'http://localhost:8080'
  }
}));

// Utility function to wait for backend to be ready
global.waitForBackend = async (url = global.testConfig.backendUrl, timeout = 30000) => {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    try {
      const fetchFn = global.fetch || require('node-fetch');
      const response = await fetchFn(`${url}/api/ca-info`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (response.status === 200 || response.status === 401) {
        // Backend is responding (200 = setup complete, 401 = needs auth)
        console.log(`✅ Backend is ready (status: ${response.status})`);
        return true;
      }
    } catch (error) {
      // Backend not ready yet
      console.log(`⏳ Backend not ready: ${error.message}`);
    }
    
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
  
  throw new Error(`Backend not ready after ${timeout}ms`);
};

// Utility function to reset backend state for tests
global.resetBackendState = async () => {
  try {
    const fetchFn = global.fetch || require('node-fetch');
    // Clear any existing sessions
    await fetchFn(`${global.testConfig.backendUrl}/api/logout`, {
      method: 'POST',
      credentials: 'include',
    });
  } catch (error) {
    // Ignore errors during cleanup
  }
};

// Setup and teardown for each test
beforeEach(async () => {
  await global.resetBackendState();
}); 