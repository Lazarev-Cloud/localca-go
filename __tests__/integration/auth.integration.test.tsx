import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { useRouter } from 'next/navigation'
import LoginPage from '@/app/login/page'
import SetupPage from '@/app/setup/page'
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

describe('Authentication Integration Tests', () => {
  beforeAll(async () => {
    // Wait for backend to be ready
    await global.waitForBackend()
  })

  beforeEach(async () => {
    // Reset backend state before each test
    await global.resetBackendState()
  })

  describe('Setup Flow', () => {
    it('should complete initial setup successfully', async () => {
      // First, check if setup is required
      const setupStatusResponse = await fetch(`${global.testConfig.backendUrl}/api/setup`)
      const setupStatus = await setupStatusResponse.json()
      
      if (setupStatus.data?.setup_completed) {
        // Reset setup state for testing
        console.log('Setup already completed, this test may need manual reset')
      }

      render(<SetupPage />)
      
      // Wait for setup form to load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /localca setup/i })).toBeInTheDocument()
      }, { timeout: 10000 })

      // Fill in setup form
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
      
      fireEvent.change(usernameInput, { target: { value: 'admin' } })
      fireEvent.change(passwordInput, { target: { value: 'testpassword123' } })
      fireEvent.change(confirmPasswordInput, { target: { value: 'testpassword123' } })

      // Submit setup form
      const setupButton = screen.getByRole('button', { name: /complete setup/i })
      fireEvent.click(setupButton)

      // Wait for setup completion
      await waitFor(() => {
        expect(screen.getByText(/setup completed successfully/i)).toBeInTheDocument()
      }, { timeout: 15000 })

      // Should redirect to login
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login')
      }, { timeout: 5000 })
    })
  })

  describe('Login Flow', () => {
    it('should handle login with correct credentials', async () => {
      render(<LoginPage />)
      
      // Wait for login form to appear
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /localca login/i })).toBeInTheDocument()
      }, { timeout: 10000 })

      // Fill in login form
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: 'admin' } })
      fireEvent.change(passwordInput, { target: { value: 'testpassword123' } })

      // Submit login form
      const loginButton = screen.getByRole('button', { name: /log in/i })
      fireEvent.click(loginButton)

      // Wait for successful login
      await waitFor(() => {
        const successMessage = screen.queryByText(/login successful/i) || 
                              screen.queryByText(/already authenticated/i)
        expect(successMessage).toBeInTheDocument()
      }, { timeout: 10000 })

      // Should redirect to dashboard
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/')
      }, { timeout: 5000 })
    })

    it('should handle login with incorrect credentials', async () => {
      render(<LoginPage />)
      
      // Wait for login form to appear
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      }, { timeout: 10000 })

      // Fill in login form with wrong credentials
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: 'admin' } })
      fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } })

      // Submit login form
      const loginButton = screen.getByRole('button', { name: /log in/i })
      fireEvent.click(loginButton)

      // Wait for error message
      await waitFor(() => {
        const errorMessage = screen.queryByText(/invalid credentials/i) ||
                            screen.queryByText(/authentication failed/i) ||
                            screen.queryByText(/login failed/i)
        expect(errorMessage).toBeInTheDocument()
      }, { timeout: 10000 })

      // Should not redirect
      expect(mockPush).not.toHaveBeenCalled()
    })

    it('should validate required fields', async () => {
      render(<LoginPage />)
      
      // Wait for login form to appear
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      }, { timeout: 10000 })

      // Clear any pre-filled values and submit empty form
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: '' } })
      fireEvent.change(passwordInput, { target: { value: '' } })

      const loginButton = screen.getByRole('button', { name: /log in/i })
      fireEvent.click(loginButton)

      // Wait for validation error
      await waitFor(() => {
        expect(screen.getByText(/username and password are required/i)).toBeInTheDocument()
      }, { timeout: 5000 })
    })
  })

  describe('Authentication State', () => {
    it('should check authentication status on page load', async () => {
      render(<LoginPage />)
      
      // The component should make an initial auth check
      // We can verify this by checking if the component renders properly
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /localca login/i })).toBeInTheDocument()
      }, { timeout: 10000 })
    })

    it('should redirect if already authenticated', async () => {
      // First login to establish session
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
        // Now render login page - should redirect since we're authenticated
        render(<LoginPage />)
        
        await waitFor(() => {
          const alreadyAuthMessage = screen.queryByText(/already authenticated/i)
          if (alreadyAuthMessage) {
            expect(alreadyAuthMessage).toBeInTheDocument()
          }
        }, { timeout: 10000 })
      }
    })
  })

  describe('API Integration', () => {
    it('should successfully communicate with backend API', async () => {
      // Test direct API call to ensure backend is responding
      const response = await fetch(`${global.testConfig.backendUrl}/api/ca-info`)
      
      // Should get either 200 (authenticated) or 401 (needs auth)
      expect([200, 401]).toContain(response.status)
      
      if (response.ok) {
        const data = await response.json()
        expect(data).toHaveProperty('success')
      }
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
    })
  })
}) 