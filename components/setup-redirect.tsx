'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useApi, ApiErrorType } from '@/hooks/use-api'

export function SetupRedirect() {
  const { error } = useApi()
  const router = useRouter()
  const [redirecting, setRedirecting] = useState(false)
  const [path, setPath] = useState<string | null>(null)
  
  useEffect(() => {
    // Clear redirection state when error changes or is cleared
    if (!error) {
      setRedirecting(false)
      setPath(null)
      return
    }
    
    // Prevent multiple redirects
    if (redirecting) return
    
    // Check if the error is a setup required error
    if (error && 
        (error.type === ApiErrorType.SETUP_REQUIRED || 
         error.setupRequired || 
         (error.status === 401 && error.message === 'Setup required'))) {
      console.log('Setup required detected, redirecting to /setup')
      setRedirecting(true)
      setPath('/setup')
      
      // Delay to avoid immediate redirect that might cause loops
      setTimeout(() => {
        router.push('/setup')
      }, 100)
    } 
    // Check if unauthorized error (not setup required) - redirect to login
    else if (error && error.status === 401 && !error.setupRequired) {
      console.log('Authentication required, redirecting to /login')
      setRedirecting(true)
      setPath('/login')
      
      // Delay to avoid immediate redirect that might cause loops
      setTimeout(() => {
        router.push('/login')
      }, 100)
    }
  }, [error, router, redirecting])
  
  // This is just a utility component, it doesn't render anything visible
  // But it can be useful for debugging to see what's happening
  if (process.env.NODE_ENV === 'development' && redirecting) {
    return (
      <div style={{ 
        position: 'fixed', 
        bottom: '10px', 
        right: '10px', 
        background: '#f0f9ff', 
        border: '1px solid #bae6fd',
        padding: '8px',
        borderRadius: '4px',
        fontSize: '12px',
        zIndex: 9999
      }}>
        Redirecting to: {path}
      </div>
    )
  }
  
  return null
} 