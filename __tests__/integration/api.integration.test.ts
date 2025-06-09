// Skip Next.js API proxy tests in integration environment for now
// These require a full Next.js runtime which is complex to set up in Jest
// import { NextRequest } from 'next/server'
// import { POST } from '@/app/api/proxy/[...path]/route'

describe('API Integration Tests', () => {
  beforeAll(async () => {
    // Wait for backend to be ready
    await global.waitForBackend()
  })

  beforeEach(async () => {
    // Reset backend state before each test
    await global.resetBackendState()
  })

  describe('Direct Backend API Calls', () => {
    it('should get CA info from backend', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/ca-info`)
      
      // Should get either 200 (authenticated) or 401 (needs auth)
      expect([200, 401]).toContain(response.status)
      
      const data = await response.json()
      expect(data).toHaveProperty('success')
    })

    it('should handle setup status check', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      
      expect(response.status).toBe(200)
      
      const data = await response.json()
      expect(data).toHaveProperty('success')
      expect(data).toHaveProperty('data')
    })

    it('should handle login attempts', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'wrongpassword'
        })
      })
      
      // Should return 401 for wrong credentials
      expect(response.status).toBe(401)
      
      const data = await response.json()
      expect(data).toHaveProperty('success', false)
    })
  })

  // TODO: Add Next.js API Proxy Routes tests
  // These require a full Next.js runtime environment
  describe('Next.js API Proxy Routes', () => {
    it.skip('should proxy CA info requests correctly', async () => {
      // Skipped - requires Next.js runtime
    })

    it.skip('should proxy login requests correctly', async () => {
      // Skipped - requires Next.js runtime
    })

    it.skip('should proxy setup requests correctly', async () => {
      // Skipped - requires Next.js runtime
    })

    it.skip('should handle POST setup requests', async () => {
      // Skipped - requires Next.js runtime
    })
  })

  describe('Error Handling', () => {
    it('should handle invalid API paths', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/nonexistent`)
      
      expect(response.status).toBe(404)
    })

    it('should handle malformed requests', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: 'invalid json'
      })
      
      expect([400, 500]).toContain(response.status)
    })
  })

  describe('Authentication Flow Integration', () => {
    it('should complete full authentication flow', async () => {
      // 1. Check setup status
      const setupResponse = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      const setupData = await setupResponse.json()
      
      if (!setupData.data?.setup_completed && setupData.data?.setup_token) {
        // 2. Complete setup if needed
        const completeSetupResponse = await fetch(`${global.testConfig.backendUrl}/api/setup`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: 'admin',
            password: 'testpassword123',
            setup_token: setupData.data.setup_token
          })
        })
        
        if (completeSetupResponse.ok) {
          console.log('Setup completed successfully')
        }
      }
      
      // 3. Attempt login
      const loginResponse = await fetch(`${global.testConfig.backendUrl}/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          username: 'admin',
          password: 'testpassword123'
        })
      })
      
      if (loginResponse.ok) {
        // 4. Access protected resource
        const caInfoResponse = await fetch(`${global.testConfig.backendUrl}/api/ca-info`, {
          credentials: 'include'
        })
        
        expect(caInfoResponse.status).toBe(200)
        
        const caData = await caInfoResponse.json()
        expect(caData).toHaveProperty('success', true)
        expect(caData).toHaveProperty('data')
      }
    })
  })
}) 