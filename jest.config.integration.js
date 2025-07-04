module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.integration.js'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
    '^public/(.*)$': '<rootDir>/public/$1',
    // Handle CSS imports (with CSS modules)
    '\\.module\\.(css|sass|scss)$': 'identity-obj-proxy',
    // Handle CSS imports (without CSS modules)
    '\\.(css|sass|scss)$': '<rootDir>/__mocks__/styleMock.js',
    // Handle static assets
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$': '<rootDir>/__mocks__/fileMock.js',
  },
  testPathIgnorePatterns: ['<rootDir>/node_modules/', '<rootDir>/.next/'],
  transform: {
    '^.+\\.(js|jsx|ts|tsx)$': ['babel-jest', { presets: ['next/babel'] }],
  },
  transformIgnorePatterns: [
    '/node_modules/',
    '^.+\\.module\\.(css|sass|scss)$',
  ],
  collectCoverageFrom: [
    '**/*.{js,jsx,ts,tsx}',
    '!**/*.d.ts',
    '!**/node_modules/**',
    '!**/.next/**',
    '!**/coverage/**',
    '!jest.config.js',
    '!jest.config.integration.js',
    '!next.config.mjs',
    '!tailwind.config.ts',
    '!postcss.config.mjs',
  ],
  testTimeout: 60000, // Increased timeout for integration tests
  testMatch: [
    '**/__tests__/integration/**/*.test.{js,jsx,ts,tsx}',
    '**/*.integration.test.{js,jsx,ts,tsx}'
  ],
  globalSetup: '<rootDir>/test-utils/global-setup.js',
  globalTeardown: '<rootDir>/test-utils/global-teardown.js',
}; 