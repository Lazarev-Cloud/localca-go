'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useApi, ApiErrorType } from '@/hooks/use-api'

export function SetupRedirect() {
  const { error } = useApi()
  const router = useRouter()
  const [redirecting, setRedirecting] = useState(false)
  
  useEffect(() => {
    // Prevent multiple redirects
    if (redirecting) return
    
    // Check if the error is a setup required error
    if (error && 
        (error.type === ApiErrorType.SETUP_REQUIRED || 
         error.setupRequired || 
         (error.status === 401 && error.message === 'Setup required'))) {
      console.log('Setup required detected, redirecting to /setup')
      setRedirecting(true)
      // Navigate to the setup page
      router.push('/setup')
    } 
    // Check if unauthorized error (not setup required) - redirect to login
    else if (error && error.status === 401 && !error.setupRequired) {
      console.log('Authentication required, redirecting to /login')
      setRedirecting(true)
      router.push('/login')
    }
  }, [error, router, redirecting])
  
  // This is just a utility component, it doesn't render anything
  return null
} 