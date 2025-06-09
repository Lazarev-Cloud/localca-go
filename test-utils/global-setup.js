const dockerSetup = require('./docker-setup');

// Use Docker-based setup functions

module.exports = async () => {
  try {
    await dockerSetup.setupTestEnvironment();
    await dockerSetup.startDockerBackend();
    console.log('🎉 Docker-based integration test environment ready!');
  } catch (error) {
    console.error('❌ Failed to setup test environment:', error);
    process.exit(1);
  }
}; 