'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import config from '@/lib/config'

export default function LoginPage() {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)
  const router = useRouter()

  // Check if already authenticated
  useEffect(() => {
    async function checkAuthStatus() {
      try {
        setChecking(true)
        // Try to access ca-info endpoint which requires authentication
        const response = await fetch('/api/ca-info', {
          credentials: 'include',
          cache: 'no-store',
          headers: {
            'Cache-Control': 'no-cache'
          }
        })
        
        // If successful, we're already authenticated
        if (response.ok) {
          setMessage('Already authenticated. Redirecting to dashboard...')
          setTimeout(() => {
            router.push('/')
          }, 1500)
          return
        }
        
        setChecking(false)
      } catch (error) {
        // If there's an error, we'll stay on the login page
        console.error('Error checking auth status:', error)
        setChecking(false)
      }
    }
    
    checkAuthStatus()
  }, [router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setMessage('')
    
    if (!username || !password) {
      setError('Username and password are required')
      return
    }

    setLoading(true)

    try {
      // Use the proxy endpoint for login
      const response = await fetch(`/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Cache-Control': 'no-cache'
        },
        body: new URLSearchParams({
          username,
          password
        }).toString(),
        credentials: 'include',
      })

      // Try to parse the response
      let data;
      try {
        data = await response.json();
      } catch (err) {
        console.error('Error parsing response:', err);
        data = { success: response.ok, message: response.ok ? 'Login successful' : 'Login failed' };
      }

      if (!response.ok) {
        setError(data.message || 'Login failed')
        return
      }

      // Success - redirect to homepage
      setMessage('Login successful! Redirecting to dashboard...')
      
      // Delay redirect to allow cookies to be set properly
      setTimeout(() => {
        router.push('/')
      }, 1500)
    } catch (err) {
      console.error('Login error:', err)
      setError('Failed to connect to the server')
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
        <div className="text-center">
          <h1 className="text-xl font-bold">Checking authentication status...</h1>
          <p className="mt-2 text-gray-500">Please wait</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-md">
        <div className="rounded-lg bg-white p-8 shadow-md">
          <h1 className="mb-6 text-2xl font-bold">LocalCA Login</h1>
          
          {error && (
            <div className="mb-4 rounded-md bg-red-50 p-4 text-red-700">
              {error}
            </div>
          )}
          
          {message && (
            <div className="mb-4 rounded-md bg-green-50 p-4 text-green-700">
              {message}
            </div>
          )}
          
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none"
                disabled={!!message}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none"
                disabled={!!message}
              />
            </div>
            
            <button
              type="submit"
              disabled={loading || !!message}
              className="w-full rounded-md bg-blue-600 py-2 px-4 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            >
              {loading ? 'Logging in...' : 'Log In'}
            </button>
          </form>

          {process.env.NODE_ENV === 'development' && (
            <div className="mt-4 text-center text-xs text-gray-500">
              <p>Backend API URL: {config.apiUrl}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
} 