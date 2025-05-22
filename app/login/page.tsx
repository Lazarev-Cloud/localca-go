'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import config from '@/lib/config'

export default function LoginPage() {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const router = useRouter()

  // Check if already authenticated
  useEffect(() => {
    async function checkAuthStatus() {
      try {
        // Try to access ca-info endpoint which requires authentication
        const response = await fetch('/api/ca-info', {
          credentials: 'include',
        })
        
        // If successful, we're already authenticated
        if (response.ok) {
          router.push('/')
        }
      } catch (error) {
        // If there's an error, we'll stay on the login page
        console.error('Error checking auth status:', error)
      }
    }
    
    checkAuthStatus()
  }, [router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password) {
      setError('Username and password are required')
      return
    }

    setLoading(true)

    try {
      // Use direct login endpoint
      const response = await fetch(`/api/login/direct`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username,
          password
        }),
        credentials: 'include',
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.message || 'Login failed')
        return
      }

      // Success - redirect to homepage
      router.push('/')
    } catch (err) {
      console.error('Login error:', err)
      setError('Failed to connect to the server')
    } finally {
      setLoading(false)
    }
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
              />
            </div>
            
            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-md bg-blue-600 py-2 px-4 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            >
              {loading ? 'Logging in...' : 'Log In'}
            </button>
          </form>

          <div className="mt-4 text-center text-xs text-gray-500">
            <p>Backend API URL: {config.apiUrl}</p>
          </div>
        </div>
      </div>
    </div>
  )
} 