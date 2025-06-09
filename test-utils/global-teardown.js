const dockerSetup = require('./docker-setup');

module.exports = async () => {
  console.log('🏁 Tearing down Docker-based integration test environment...');
  
  try {
    await dockerSetup.stopDockerBackend();
    await dockerSetup.cleanupTestData();
    console.log('✅ Docker-based integration test environment cleaned up');
  } catch (error) {
    console.error('❌ Error during teardown:', error);
  }
}; 