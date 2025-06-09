import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { useRouter } from 'next/navigation'
import LoginPage from '@/app/login/page'
import SetupPage from '@/app/setup/page'
import '@testing-library/jest-dom'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

// Mock fetch globally
global.fetch = jest.fn()

const mockPush = jest.fn()
const mockRouter = {
  push: mockPush,
  replace: jest.fn(),
  prefetch: jest.fn(),
}

beforeEach(() => {
  ;(useRouter as jest.Mock).mockReturnValue(mockRouter)
  ;(fetch as jest.Mock).mockClear()
  mockPush.mockClear()
})

describe('Authentication Flow', () => {
  describe('LoginPage', () => {
    it('renders login form with pre-filled credentials', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /localca login/i })).toBeInTheDocument()
      })
      
      expect(screen.getByLabelText(/username/i)).toHaveValue('admin')
      expect(screen.getByLabelText(/password/i)).toHaveValue('12345678')
      expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
    })

    it('shows validation error for empty fields', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      })

      // Clear the pre-filled values
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: '' } })
      fireEvent.change(passwordInput, { target: { value: '' } })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/username and password are required/i)).toBeInTheDocument()
      })
    })

    it('handles successful login', async () => {
      // Mock the initial auth check to fail, then successful login
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            message: 'Login successful'
          })
        })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/login successful/i)).toBeInTheDocument()
      })

      // Should redirect after delay
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/')
      }, { timeout: 2000 })
    })

    it('handles login failure', async () => {
      // Mock the initial auth check to fail, then failed login
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockResolvedValueOnce({
          ok: false,
          json: async () => ({
            success: false,
            message: 'Invalid credentials'
          })
        })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
      })

      expect(mockPush).not.toHaveBeenCalled()
    })

    it('handles network errors', async () => {
      // Mock the initial auth check to fail, then network error
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockRejectedValueOnce(new Error('Network error'))

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/failed to connect to the server/i)).toBeInTheDocument()
      })
    })

    it('checks authentication status on mount', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/proxy/api/ca-info', expect.any(Object))
      })
    })

    it('redirects if already authenticated', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      render(<LoginPage />)
      
      await waitFor(() => {
        expect(screen.getByText(/already authenticated/i)).toBeInTheDocument()
      })

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/')
      }, { timeout: 2000 })
    })

    it('disables form during loading', async () => {
      // Mock the initial auth check to fail, then slow login
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockImplementationOnce(() => 
          new Promise(resolve => setTimeout(resolve, 1000))
        )

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /logging in/i })).toBeDisabled()
        expect(screen.getByLabelText(/username/i)).toBeDisabled()
        expect(screen.getByLabelText(/password/i)).toBeDisabled()
      })
    })
  })

  describe('SetupPage', () => {
    it('renders setup form', async () => {
      // Mock the initial setup check
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            setup_completed: false,
            setup_required: true,
            setup_token: 'test_token_123'
          }
        })
      })

      render(<SetupPage />)
      
      // Wait for the form to appear after setup check
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /localca initial setup/i })).toBeInTheDocument()
      })
      
      expect(screen.getByLabelText(/username/i)).toHaveValue('admin')
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/setup token/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /complete setup/i })).toBeInTheDocument()
    })

    it('loads setup status on mount', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            setup_completed: false,
            setup_required: true,
            setup_token: 'test_token_123'
          }
        })
      })

      render(<SetupPage />)
      
      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/setup', expect.any(Object))
      })

      await waitFor(() => {
        expect(screen.getByDisplayValue('test_token_123')).toBeInTheDocument()
      })
    })

    it('redirects if setup already completed', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            setup_completed: true,
            setup_required: false
          }
        })
      })

      render(<SetupPage />)
      
      await waitFor(() => {
        expect(screen.getByText(/setup already completed/i)).toBeInTheDocument()
      })

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login')
      }, { timeout: 2500 })
    })

    it('validates required fields', async () => {
      // Mock the initial setup check
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            setup_completed: false,
            setup_required: true,
            setup_token: 'test_token_123'
          }
        })
      })

      render(<SetupPage />)
      
      // Wait for the form to appear after setup check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /complete setup/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/all fields are required/i)).toBeInTheDocument()
      })
    })

    it('validates password confirmation', async () => {
      // Mock the initial setup check
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            setup_completed: false,
            setup_required: true,
            setup_token: 'test_token_123'
          }
        })
      })

      render(<SetupPage />)
      
      // Wait for the form to appear after setup check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /complete setup/i })).toBeInTheDocument()
      })
      
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'password123' } })
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'different' } })
      fireEvent.change(screen.getByLabelText(/setup token/i), { target: { value: 'token' } })
      
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument()
      })
    })

    it('handles successful setup', async () => {
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            data: { setup_token: 'test_token' }
          })
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            message: 'Setup completed successfully'
          })
        })

      render(<SetupPage />)
      
      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByDisplayValue('test_token')).toBeInTheDocument()
      })

      // Fill form
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'password123' } })
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'password123' } })
      
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/setup completed successfully/i)).toBeInTheDocument()
      })

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login')
      }, { timeout: 2500 })
    })

    it('handles setup failure', async () => {
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            data: { setup_token: 'test_token' }
          })
        })
        .mockResolvedValueOnce({
          ok: false,
          json: async () => ({
            success: false,
            message: 'Invalid setup token'
          })
        })

      render(<SetupPage />)
      
      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByDisplayValue('test_token')).toBeInTheDocument()
      })

      // Fill form
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'password123' } })
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'password123' } })
      
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/invalid setup token/i)).toBeInTheDocument()
      })

      expect(mockPush).not.toHaveBeenCalled()
    })

    it('disables form during submission', async () => {
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            data: { setup_token: 'test_token' }
          })
        })
        .mockImplementationOnce(() => 
          new Promise(resolve => setTimeout(resolve, 1000))
        )

      render(<SetupPage />)
      
      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByDisplayValue('test_token')).toBeInTheDocument()
      })

      // Fill form
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'password123' } })
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'password123' } })
      
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /setting up/i })).toBeDisabled()
      })
    })
  })

  describe('Form Interactions', () => {
    it('allows typing in all form fields', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      })
      
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      
      fireEvent.change(usernameInput, { target: { value: 'testuser' } })
      fireEvent.change(passwordInput, { target: { value: 'testpass' } })
      
      expect(usernameInput).toHaveValue('testuser')
      expect(passwordInput).toHaveValue('testpass')
    })

    it('submits form on Enter key', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true })
        })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
      })
      
      const passwordInput = screen.getByLabelText(/password/i)
      fireEvent.keyDown(passwordInput, { key: 'Enter', code: 'Enter' })
      
      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/proxy/api/login', expect.any(Object))
      })
    })

    it('clears error messages when typing', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      })
      
      // Trigger error
      fireEvent.change(screen.getByLabelText(/username/i), { target: { value: '' } })
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(screen.getByText(/username and password are required/i)).toBeInTheDocument()
      })

      // Start typing
      fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'a' } })
      
      // Error should be cleared
      expect(screen.queryByText(/username and password are required/i)).not.toBeInTheDocument()
    })
  })

  describe('API Integration', () => {
    it('sends correct login request format', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true })
        })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/proxy/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Cache-Control': 'no-cache'
          },
          body: JSON.stringify({
            username: 'admin',
            password: '12345678'
          }),
          credentials: 'include'
        })
      })
    })

    it('sends correct setup request format', async () => {
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({
            success: true,
            data: { setup_token: 'test_token' }
          })
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true })
        })

      render(<SetupPage />)
      
      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByDisplayValue('test_token')).toBeInTheDocument()
      })

      // Fill and submit form
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'testpass' } })
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'testpass' } })
      fireEvent.click(screen.getByRole('button', { name: /complete setup/i }))
      
      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/setup', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Cache-Control': 'no-cache'
          },
          body: JSON.stringify({
            username: 'admin',
            password: 'testpass',
            confirm_password: 'testpass',
            setup_token: 'test_token'
          })
        })
      })
    })

    it('handles malformed JSON responses', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 401
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => { throw new Error('Invalid JSON') }
        })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByRole('button', { name: /log in/i }))
      
      // Should not crash and should handle gracefully
      await waitFor(() => {
        expect(screen.getByText(/login successful/i)).toBeInTheDocument()
      })
    })
  })

  describe('Accessibility', () => {
    it('has proper form labels', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      })
      
      expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    })

    it('has proper form structure', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: { setup_token: 'test_token' }
        })
      })

      render(<SetupPage />)
      
      // Wait for the form to appear
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /complete setup/i })).toBeInTheDocument()
      })
      
      const inputs = screen.getAllByRole('textbox')
      expect(inputs.length).toBeGreaterThan(0)
      
      const passwordInputs = screen.getAllByLabelText(/password/i)
      expect(passwordInputs.length).toBe(2)
    })

    it('supports keyboard navigation', async () => {
      // Mock the initial auth check to fail so we see the login form
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401
      })

      render(<LoginPage />)
      
      // Wait for the form to appear after auth check
      await waitFor(() => {
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
      })
      
      const usernameInput = screen.getByLabelText(/username/i)
      const passwordInput = screen.getByLabelText(/password/i)
      const submitButton = screen.getByRole('button', { name: /log in/i })
      
      usernameInput.focus()
      expect(document.activeElement).toBe(usernameInput)
      
      fireEvent.keyDown(usernameInput, { key: 'Tab' })
      expect(document.activeElement).toBe(passwordInput)
      
      fireEvent.keyDown(passwordInput, { key: 'Tab' })
      expect(document.activeElement).toBe(submitButton)
    })
  })
}) 