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
  })
}) 