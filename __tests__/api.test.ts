import { NextRequest } from 'next/server'
import { POST } from '@/app/api/proxy/[...path]/route'

// Mock fetch for backend calls
global.fetch = jest.fn()

describe('API Proxy Routes', () => {
  beforeEach(() => {
    ;(fetch as jest.Mock).mockClear()
  })

  describe('Login API Proxy', () => {
    it('forwards login requests correctly', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json',
          'Set-Cookie': 'session=test_session_token; HttpOnly; Secure'
        }),
        json: async () => ({
          success: true,
          message: 'Login successful'
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'Mozilla/5.0 Test Browser'
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'testpass'
        })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(200)
      
      const data = await response.json()
      expect(data.success).toBe(true)
      expect(data.message).toBe('Login successful')

      // Verify backend was called correctly
      expect(fetch).toHaveBeenCalledWith('http://backend:8080/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'Mozilla/5.0 Test Browser',
          'X-Forwarded-For': expect.any(String),
          'X-Real-IP': expect.any(String)
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'testpass'
        })
      })
    })

    it('handles login failures correctly', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => ({
          success: false,
          message: 'Invalid credentials'
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'wrongpass'
        })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(401)
      
      const data = await response.json()
      expect(data.success).toBe(false)
      expect(data.message).toBe('Invalid credentials')
    })

    it('forwards cookies from backend', async () => {
      const sessionCookie = 'session=abc123; HttpOnly; Secure; Path=/'
      
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json',
          'Set-Cookie': sessionCookie
        }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'testpass' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.headers.get('Set-Cookie')).toBe(sessionCookie)
    })
  })

  describe('Setup API Proxy', () => {
    it('forwards setup requests correctly', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => ({
          success: true,
          message: 'Setup completed successfully'
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/setup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'newpass',
          setup_token: 'valid_token'
        })
      })

      const response = await POST(request, { params: { path: ['api', 'setup'] } })
      
      expect(response.status).toBe(200)
      
      const data = await response.json()
      expect(data.success).toBe(true)
      expect(data.message).toBe('Setup completed successfully')
    })

    it('handles GET setup status requests', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => ({
          success: true,
          data: {
            setup_completed: false,
            setup_required: true,
            setup_token: 'test_token_123'
          }
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/setup', {
        method: 'GET'
      })

      const response = await POST(request, { params: { path: ['api', 'setup'] } })
      
      expect(response.status).toBe(200)
      
      const data = await response.json()
      expect(data.success).toBe(true)
      expect(data.data.setup_token).toBe('test_token_123')
    })
  })

  describe('Certificate API Proxy', () => {
    it('forwards certificate requests with authentication', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => ({
          success: true,
          data: {
            certificates: []
          }
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/certificates', {
        method: 'GET',
        headers: {
          'Cookie': 'session=valid_session_token'
        }
      })

      const response = await POST(request, { params: { path: ['api', 'certificates'] } })
      
      expect(response.status).toBe(200)
      
      // Verify session cookie was forwarded
      expect(fetch).toHaveBeenCalledWith('http://backend:8080/api/certificates', {
        method: 'GET',
        headers: {
          'Cookie': 'session=valid_session_token',
          'X-Forwarded-For': expect.any(String),
          'X-Real-IP': expect.any(String)
        }
      })
    })

    it('handles certificate creation requests', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 201,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => ({
          success: true,
          message: 'Certificate created successfully',
          data: {
            id: 'cert_123',
            common_name: 'test.local'
          }
        })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/certificates', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cookie': 'session=valid_session_token'
        },
        body: JSON.stringify({
          common_name: 'test.local',
          certificate_type: 'server',
          validity_days: 365
        })
      })

      const response = await POST(request, { params: { path: ['api', 'certificates'] } })
      
      expect(response.status).toBe(201)
      
      const data = await response.json()
      expect(data.success).toBe(true)
      expect(data.data.common_name).toBe('test.local')
    })
  })

  describe('Error Handling', () => {
    it('handles backend connection errors', async () => {
      ;(fetch as jest.Mock).mockRejectedValueOnce(new Error('ECONNREFUSED'))

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'test' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(502)
      
      const data = await response.json()
      expect(data.success).toBe(false)
      expect(data.message).toContain('Backend service unavailable')
    })

    it('handles timeout errors', async () => {
      ;(fetch as jest.Mock).mockImplementationOnce(() => 
        new Promise((_, reject) => 
          setTimeout(() => reject(new Error('Request timeout')), 100)
        )
      )

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'test' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(502)
    })

    it('handles malformed JSON from backend', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json'
        }),
        json: async () => { throw new Error('Invalid JSON') },
        text: async () => 'Invalid JSON response'
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'test' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(502)
      
      const data = await response.json()
      expect(data.success).toBe(false)
      expect(data.message).toContain('Invalid response from backend')
    })
  })

  describe('Request Validation', () => {
    it('validates required fields for login', async () => {
      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: '' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      // Should still forward to backend for validation
      expect(fetch).toHaveBeenCalled()
    })

    it('handles different content types', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: 'username=admin&password=test'
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      expect(response.status).toBe(200)
      
      // Verify content type was preserved
      expect(fetch).toHaveBeenCalledWith(expect.any(String), {
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/x-www-form-urlencoded'
        }),
        body: 'username=admin&password=test'
      })
    })
  })

  describe('Security Headers', () => {
    it('adds security headers to responses', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'test' })
      })

      const response = await POST(request, { params: { path: ['api', 'login'] } })
      
      // Check for security headers
      expect(response.headers.get('X-Content-Type-Options')).toBe('nosniff')
      expect(response.headers.get('X-Frame-Options')).toBe('DENY')
      expect(response.headers.get('X-XSS-Protection')).toBe('1; mode=block')
    })

    it('preserves CORS headers from backend', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE'
        }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/certificates', {
        method: 'GET'
      })

      const response = await POST(request, { params: { path: ['api', 'certificates'] } })
      
      expect(response.headers.get('Access-Control-Allow-Origin')).toBe('*')
      expect(response.headers.get('Access-Control-Allow-Methods')).toBe('GET, POST, PUT, DELETE')
    })
  })

  describe('Path Handling', () => {
    it('handles nested API paths correctly', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/certificates/123/download', {
        method: 'GET'
      })

      const response = await POST(request, { params: { path: ['api', 'certificates', '123', 'download'] } })
      
      expect(response.status).toBe(200)
      
      // Verify correct backend URL was called
      expect(fetch).toHaveBeenCalledWith('http://backend:8080/api/certificates/123/download', expect.any(Object))
    })

    it('handles query parameters correctly', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => ({ success: true })
      })

      const request = new NextRequest('http://localhost:3000/api/proxy/api/certificates?page=1&limit=10', {
        method: 'GET'
      })

      const response = await POST(request, { params: { path: ['api', 'certificates'] } })
      
      expect(response.status).toBe(200)
      
      // Verify query parameters were preserved
      expect(fetch).toHaveBeenCalledWith('http://backend:8080/api/certificates?page=1&limit=10', expect.any(Object))
    })
  })
}) 