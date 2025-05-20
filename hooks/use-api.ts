"use client"

import { useState } from 'react'
import config from '@/lib/config'

type ApiResponse<T> = {
  success: boolean
  message: string
  data?: T
}

export function useApi() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchApi = async <T>(
    endpoint: string, 
    options?: RequestInit
  ): Promise<ApiResponse<T>> => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      })

      if (!response.ok) {
        throw new Error(`API request failed with status ${response.status}`)
      }

      const data = await response.json()
      return data as ApiResponse<T>
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred'
      setError(errorMessage)
      return {
        success: false,
        message: errorMessage,
      }
    } finally {
      setLoading(false)
    }
  }

  const postFormData = async <T>(
    endpoint: string,
    formData: FormData
  ): Promise<ApiResponse<T>> => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api${endpoint}`, {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        throw new Error(`API request failed with status ${response.status}`)
      }

      const data = await response.json()
      return data as ApiResponse<T>
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred'
      setError(errorMessage)
      return {
        success: false,
        message: errorMessage,
      }
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    fetchApi,
    postFormData,
  }
} 