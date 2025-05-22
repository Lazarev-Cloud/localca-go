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
  const [setupCompleted, setSetupCompleted] = useState(false)
  const router = useRouter()

  // Get the setup token from backend
  useEffect(() => {
    // Check setup status and get token
    async function checkSetupStatus() {
      try {
        const response = await fetch('/api/setup', {
          cache: 'no-store', // Ensure fresh data
          headers: {
            'Cache-Control': 'no-cache'
          }
        })
        const data = await response.json()
        
        if (data.success && data.data) {
          // If setup is already completed, redirect to login
          if (data.data.setup_completed) {
            setSetupCompleted(true)
            setSetupInfo('Setup already completed. Redirecting to login page...')
            
            // Delay redirection to give user time to read the message
            setTimeout(() => {
              router.push('/login')
            }, 2000)
          } else if (data.data.setup_token) {
            // Auto-fill the setup token
            setSetupToken(data.data.setup_token)
            setSetupInfo('Setup token has been automatically loaded from the backend.')
          }
        }
      } catch (error) {
        console.error('Error checking setup status:', error)
        setError('Error checking setup status. Please try again.')
        setSetupInfo('Could not automatically load setup token. Please check your backend logs for the setup token.')
      }
    }
    
    checkSetupStatus()
  }, [router])

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
          'Cache-Control': 'no-cache'
        },
        body: JSON.stringify({
          username,
          password,
          confirm_password: confirmPassword,
          setup_token: setupToken
        }),
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.message || 'Failed to complete setup')
        return
      }

      // Success - redirect to login page
      setSetupCompleted(true)
      setSetupInfo('Setup completed successfully! Redirecting to login page...')
      
      // Delay redirection to give user time to read the message
      setTimeout(() => {
        router.push('/login')
      }, 2000)
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
          
          {setupCompleted ? (
            <div className="mb-4 rounded-md bg-green-50 p-4 text-green-700">
              {setupInfo}
            </div>
          ) : (
            <>
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
            </>
          )}

          <div className="mt-4 text-center text-xs text-gray-500">
            <p>Backend API URL: {config.apiUrl}</p>
          </div>
        </div>
      </div>
    </div>
  )
} 