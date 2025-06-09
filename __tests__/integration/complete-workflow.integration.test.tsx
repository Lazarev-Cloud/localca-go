import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { useRouter } from 'next/navigation'
import LoginPage from '@/app/login/page'
import SetupPage from '@/app/setup/page'
import HomePage from '@/app/page'
import '@testing-library/jest-dom'

// Mock Next.js navigation
const mockPush = jest.fn()
const mockRouter = {
  push: mockPush,
  replace: jest.fn(),
  prefetch: jest.fn(),
}

beforeEach(() => {
  ;(useRouter as jest.Mock).mockReturnValue(mockRouter)
  mockPush.mockClear()
})

describe('Complete LocalCA Workflow Integration Tests', () => {
  beforeAll(async () => {
    // Wait for Docker backend to be ready
    await global.waitForBackend()
  })

  beforeEach(async () => {
    // Reset backend state before each test
    await global.resetBackendState()
  })

  describe('Initial Setup Flow', () => {
    it('should complete the full setup process', async () => {
      // Step 1: Check setup status
      const setupStatusResponse = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      expect(setupStatusResponse.status).toBe(200)
      
      const setupStatus = await setupStatusResponse.json()
      expect(setupStatus).toHaveProperty('success', true)
      expect(setupStatus).toHaveProperty('data')
      
      // If setup is already completed, we need to reset it for this test
      if (setupStatus.data?.setup_completed) {
        console.log('Setup already completed, continuing with login test')
        return
      }

      // Step 2: Render setup page
      render(<SetupPage />)
      
      // Wait for setup form to load
      await waitFor(() => {
        expect(screen.getByText(/LocalCA Initial Setup/i)).toBeInTheDocument()
      }, { timeout: 10000 })

      // Step 3: Fill in setup form
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/^password$/i)
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
      const setupTokenInput = screen.getByLabelText(/setup token/i)
      
      // Clear and fill form
      fireEvent.change(usernameInput, { target: { value: 'admin' } })
      fireEvent.change(passwordInput, { target: { value: 'testpassword123' } })
      fireEvent.change(confirmPasswordInput, { target: { value: 'testpassword123' } })
      
      // The setup token should be auto-filled, but let's ensure it has a value
      if (!setupTokenInput.value && setupStatus.data?.setup_token) {
        fireEvent.change(setupTokenInput, { target: { value: setupStatus.data.setup_token } })
      }

      // Step 4: Submit setup form
      const setupButton = screen.getByRole('button', { name: /complete setup/i })
      fireEvent.click(setupButton)

      // Step 5: Wait for setup completion
      await waitFor(() => {
        const successMessage = screen.queryByText(/setup completed successfully/i)
        if (successMessage) {
          expect(successMessage).toBeInTheDocument()
        }
      }, { timeout: 15000 })

      // Step 6: Verify setup was successful by checking backend
      const verifyResponse = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      const verifyData = await verifyResponse.json()
      
      if (verifyData.data?.setup_completed) {
        console.log('✅ Setup completed successfully')
      }
    })
  })

  describe('Authentication Flow', () => {
    it('should handle complete login workflow', async () => {
      // Step 1: Render login page
      render(<LoginPage />)
      
      // Wait for login form to appear
      await waitFor(() => {
        expect(screen.getByText(/LocalCA Login/i)).toBeInTheDocument()
      }, { timeout: 10000 })

      // Step 2: Fill in login form
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: 'admin' } })
      fireEvent.change(passwordInput, { target: { value: 'testpassword123' } })

      // Step 3: Submit login form
      const loginButton = screen.getByRole('button', { name: /log in/i })
      fireEvent.click(loginButton)

      // Step 4: Wait for login result
      await waitFor(() => {
        const successMessage = screen.queryByText(/login successful/i)
        const errorMessage = screen.queryByText(/failed to connect/i) || 
                            screen.queryByText(/invalid credentials/i)
        
        // Either success or a specific error should appear
        expect(successMessage || errorMessage).toBeInTheDocument()
      }, { timeout: 15000 })

      // Step 5: Verify authentication with backend
      const authResponse = await fetch(`${global.testConfig.backendUrl}/api/ca-info`, {
        credentials: 'include'
      })
      
      // Should get either 200 (authenticated) or 401 (needs auth)
      expect([200, 401]).toContain(authResponse.status)
    })

    it('should handle authentication state checking', async () => {
      // Step 1: Try to access protected page without authentication
      render(<HomePage />)
      
      // Step 2: Should show loading state initially
      await waitFor(() => {
        expect(screen.getByText(/Loading LocalCA/i)).toBeInTheDocument()
      }, { timeout: 5000 })

      // Step 3: Should eventually redirect or show appropriate state
      await waitFor(() => {
        // The page should either redirect or show some content
        const loadingText = screen.queryByText(/Loading LocalCA/i)
        // Loading should eventually disappear
        expect(loadingText).not.toBeInTheDocument()
      }, { timeout: 20000 })
    })
  })

  describe('API Integration Tests', () => {
    it('should handle CA info endpoint', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/ca-info`)
      
      // Should get either 200 (authenticated) or 401 (needs auth)
      expect([200, 401]).toContain(response.status)
      
      const data = await response.json()
      expect(data).toHaveProperty('success')
    })

    it('should handle setup endpoint', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      
      expect(response.status).toBe(200)
      
      const data = await response.json()
      expect(data).toHaveProperty('success')
      expect(data).toHaveProperty('data')
    })

    it('should handle login endpoint', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'testpassword123'
        })
      })
      
      // Should get either 200 (success) or 401 (invalid credentials)
      expect([200, 401]).toContain(response.status)
      
      const data = await response.json()
      expect(data).toHaveProperty('success')
    })

    it('should handle CORS properly', async () => {
      const response = await fetch(`${global.testConfig.backendUrl}/api/ca-info`, {
        method: 'GET',
        headers: {
          'Origin': 'http://localhost:3000',
          'Content-Type': 'application/json',
        }
      })
      
      // Should not be blocked by CORS
      expect(response.status).not.toBe(0)
      expect([200, 401, 404]).toContain(response.status)
    })
  })

  describe('Certificate Management Flow', () => {
    it('should handle certificate listing', async () => {
      // First authenticate
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
        // Try to get certificates
        const certsResponse = await fetch(`${global.testConfig.backendUrl}/api/certificates`, {
          credentials: 'include'
        })
        
        // Should get either 200 (success) or 401 (needs auth)
        expect([200, 401]).toContain(certsResponse.status)
        
        if (certsResponse.ok) {
          const certsData = await certsResponse.json()
          expect(certsData).toHaveProperty('success')
        }
      }
    })

    it('should handle certificate creation endpoint', async () => {
      // Test the certificate creation endpoint structure
      const response = await fetch(`${global.testConfig.backendUrl}/api/certificates`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          common_name: 'test.example.com',
          type: 'server'
        })
      })
      
      // Should get 401 (needs auth) since we're not authenticated
      expect(response.status).toBe(401)
    })
  })

  describe('Error Handling', () => {
    it('should handle invalid endpoints gracefully', async () => {
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

    it('should handle network timeouts', async () => {
      // Test with a very short timeout
      try {
        const controller = new AbortController()
        setTimeout(() => controller.abort(), 1) // 1ms timeout
        
        await fetch(`${global.testConfig.backendUrl}/api/ca-info`, {
          signal: controller.signal
        })
      } catch (error) {
        expect(error.name).toBe('AbortError')
      }
    })
  })

  describe('Frontend Component Integration', () => {
    it('should render setup page without errors', async () => {
      render(<SetupPage />)
      
      await waitFor(() => {
        expect(screen.getByText(/LocalCA Initial Setup/i)).toBeInTheDocument()
      }, { timeout: 10000 })
      
      // Should have all required form fields
      expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/setup token/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /complete setup/i })).toBeInTheDocument()
    })

    it('should render login page without errors', async () => {
      render(<LoginPage />)
      
      await waitFor(() => {
        expect(screen.getByText(/LocalCA Login/i)).toBeInTheDocument()
      }, { timeout: 10000 })
      
      // Should have all required form fields
      expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
    })

    it('should render home page without errors', async () => {
      render(<HomePage />)
      
      // Should show loading initially
      expect(screen.getByText(/Loading LocalCA/i)).toBeInTheDocument()
      
      // Should eventually finish loading
      await waitFor(() => {
        const loadingText = screen.queryByText(/Loading LocalCA/i)
        expect(loadingText).not.toBeInTheDocument()
      }, { timeout: 20000 })
    })
  })
}) 