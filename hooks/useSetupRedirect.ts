'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export function useSetupRedirect(apiError: any) {
  const router = useRouter()

  useEffect(() => {
    // Check if the error indicates that setup is required
    if (apiError && 
      (apiError.setupRequired || 
       (apiError.status === 401 && apiError.message === 'Setup required'))) {
      // Redirect to setup page
      router.push('/setup')
    }
  }, [apiError, router])
} 