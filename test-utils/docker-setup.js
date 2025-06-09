const { spawn, exec } = require('child_process');
const { promisify } = require('util');
const fs = require('fs').promises;
const path = require('path');

const execAsync = promisify(exec);
const sleep = promisify(setTimeout);

// Configuration
const BACKEND_PORT = 8080;
const BACKEND_URL = `http://localhost:${BACKEND_PORT}`;
const MAX_STARTUP_TIME = 120000; // 2 minutes for Docker
const HEALTH_CHECK_INTERVAL = 3000; // 3 seconds

let dockerProcess = null;

async function checkBackendHealth() {
  try {
    const fetchFn = global.fetch || require('node-fetch');
    const response = await fetchFn(`${BACKEND_URL}/api/ca-info`, {
      timeout: 5000
    });
    return response.status === 200 || response.status === 401; // Both indicate backend is running
  } catch (error) {
    return false;
  }
}

async function waitForBackend() {
  const startTime = Date.now();
  
  while (Date.now() - startTime < MAX_STARTUP_TIME) {
    if (await checkBackendHealth()) {
      console.log('✅ Backend is ready for testing');
      return true;
    }
    
    console.log('⏳ Waiting for backend to start...');
    await sleep(HEALTH_CHECK_INTERVAL);
  }
  
  throw new Error(`Backend failed to start within ${MAX_STARTUP_TIME}ms`);
}

async function setupTestEnvironment() {
  console.log('🚀 Setting up Docker-based integration test environment...');
  
  // Ensure test data directory exists
  const testDataDir = path.join(process.cwd(), 'test-data');
  try {
    await fs.access(testDataDir);
  } catch (error) {
    await fs.mkdir(testDataDir, { recursive: true });
    console.log('📁 Created test-data directory');
  }
  
  // Create CA key password file if it doesn't exist
  const caKeyFile = path.join(testDataDir, 'cakey.txt');
  try {
    await fs.access(caKeyFile);
  } catch (error) {
    await fs.writeFile(caKeyFile, 'test-ca-password-123');
    console.log('🔑 Created CA key password file');
  }
  
  console.log('⚙️  Test environment configured');
}

async function startDockerBackend() {
  console.log('🐳 Starting Docker backend...');
  
  // Check if backend is already running
  if (await checkBackendHealth()) {
    console.log('ℹ️  Backend already running, using existing instance');
    return;
  }
  
  // Stop any existing test containers
  try {
    console.log('🧹 Cleaning up existing test containers...');
    await execAsync('docker-compose -f docker-compose.test.yml down --remove-orphans');
  } catch (error) {
    // Ignore errors if containers don't exist
  }
  
  // Build and start the test backend
  console.log('🔨 Building and starting test backend...');
  dockerProcess = spawn('docker-compose', ['-f', 'docker-compose.test.yml', 'up', '--build'], {
    stdio: 'pipe',
    env: { ...process.env },
    detached: false,
  });
  
  // Handle Docker output
  dockerProcess.stdout.on('data', (data) => {
    const output = data.toString().trim();
    if (output) {
      console.log(`[Docker] ${output}`);
    }
  });
  
  dockerProcess.stderr.on('data', (data) => {
    const output = data.toString().trim();
    if (output && !output.includes('WARNING')) {
      console.error(`[Docker Error] ${output}`);
    }
  });
  
  dockerProcess.on('error', (error) => {
    console.error('Failed to start Docker backend:', error);
  });
  
  // Store process for cleanup
  global.__DOCKER_PROCESS__ = dockerProcess;
  
  // Wait for backend to be ready
  await waitForBackend();
}

async function stopDockerBackend() {
  console.log('🛑 Stopping Docker backend...');
  
  try {
    // Stop Docker Compose services
    await execAsync('docker-compose -f docker-compose.test.yml down --remove-orphans');
    console.log('✅ Docker services stopped');
  } catch (error) {
    console.error('⚠️  Error stopping Docker services:', error.message);
  }
  
  // Kill the docker-compose process if it's still running
  const process = global.__DOCKER_PROCESS__;
  if (process && !process.killed) {
    try {
      process.kill('SIGTERM');
      
      // Wait a bit for graceful shutdown
      await sleep(3000);
      
      if (!process.killed) {
        process.kill('SIGKILL');
      }
    } catch (error) {
      console.error('⚠️  Error killing Docker process:', error.message);
    }
  }
}

async function cleanupTestData() {
  console.log('🧹 Cleaning up test data...');
  
  try {
    const testDataDir = path.join(process.cwd(), 'test-data');
    
    // Remove test-specific files but keep the directory structure
    const filesToRemove = [
      'ca.crt',
      'ca.key',
      'auth.json',
      'certificates.json',
      'revoked.json',
    ];
    
    for (const file of filesToRemove) {
      const filePath = path.join(testDataDir, file);
      try {
        await fs.unlink(filePath);
        console.log(`🗑️  Removed ${file}`);
      } catch (error) {
        // File might not exist, which is fine
      }
    }
    
    // Clean up certificates directory
    const certsDir = path.join(testDataDir, 'certificates');
    try {
      const files = await fs.readdir(certsDir);
      for (const file of files) {
        await fs.unlink(path.join(certsDir, file));
      }
      console.log('🗑️  Cleaned certificates directory');
    } catch (error) {
      // Directory might not exist
    }
    
  } catch (error) {
    console.error('⚠️  Error during cleanup:', error.message);
  }
}

module.exports = {
  setupTestEnvironment,
  startDockerBackend,
  stopDockerBackend,
  cleanupTestData,
  waitForBackend,
  checkBackendHealth,
  BACKEND_URL
}; 