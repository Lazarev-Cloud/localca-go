'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import config from '@/lib/config'

export default function SetupPage() {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [setupToken, setSetupToken] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [setupInfo, setSetupInfo] = useState('')
  const router = useRouter()

  // Get the setup token from backend logs
  useEffect(() => {
    // This will display info about what to do
    setSetupInfo('Please check your backend logs for the setup token. You can find it by running: docker logs localca-backend | grep "Setup token"')
    
    // Check if setup is already completed
    async function checkSetupStatus() {
      try {
        const response = await fetch('/api/setup')
        const data = await response.json()
        
        // If setup is already completed, redirect to login
        if (data.data && data.data.setup_completed) {
          console.log('Setup already completed, redirecting to login')
          router.push('/login')
        }
      } catch (error) {
        console.error('Error checking setup status:', error)
      }
    }
    
    checkSetupStatus()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password || !confirmPassword || !setupToken) {
      setError('All fields are required')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setLoading(true)

    try {
      // Use Next.js API route to avoid CORS issues
      const response = await fetch(`/api/setup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username,
          password,
          confirm_password: confirmPassword,
          setup_token: setupToken
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        setError(data.message || 'Failed to complete setup')
        return
      }

      // Success - redirect to homepage
      alert('Setup completed successfully! Redirecting to homepage.')
      router.push('/')
    } catch (err) {
      console.error('Setup error:', err)
      setError('Failed to connect to the server')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-md">
        <div className="rounded-lg bg-white p-8 shadow-md">
          <h1 className="mb-6 text-2xl font-bold">LocalCA Initial Setup</h1>
          
          {setupInfo && (
            <div className="mb-4 rounded-md bg-blue-50 p-4 text-blue-700">
              {setupInfo}
            </div>
          )}
          
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
            
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Confirm Password
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Setup Token
              </label>
              <input
                type="text"
                value={setupToken}
                onChange={(e) => setSetupToken(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none"
              />
              <p className="mt-1 text-xs text-gray-500">
                This is the token displayed in the backend logs
              </p>
            </div>
            
            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-md bg-blue-600 py-2 px-4 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            >
              {loading ? 'Setting up...' : 'Complete Setup'}
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