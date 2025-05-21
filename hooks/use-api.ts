"use client"

import { useState } from 'react'

// Define API response type
export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data?: T
}

// Define error types
export enum ApiErrorType {
  NETWORK = 'network',
  SERVER = 'server',
  TIMEOUT = 'timeout',
  UNKNOWN = 'unknown'
}

export interface ApiError {
  type: ApiErrorType
  message: string
  status?: number
}

export function useApi() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ApiError | null>(null)

  const fetchApi = async <T>(
    endpoint: string, 
    options?: RequestInit,
    retries = 2
  ): Promise<ApiResponse<T>> => {
    setLoading(true)
    setError(null)
    
    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 10000) // 10 second timeout
      
      const response = await fetch(`/api${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        signal: controller.signal,
      })
      
      clearTimeout(timeoutId)

      if (!response.ok) {
        // Handle different error status codes
        let errorType = ApiErrorType.SERVER
        let errorMessage = `API request failed with status ${response.status}`
        
        try {
          // Try to parse error message from response
          const errorData = await response.json()
          if (errorData && errorData.message) {
            errorMessage = errorData.message
          }
        } catch (e) {
          // Ignore JSON parse error and use default message
        }
        
        const apiError: ApiError = {
          type: errorType,
          message: errorMessage,
          status: response.status
        }
        
        setError(apiError)
        
        // Retry on server errors (5xx) if retries left
        if (response.status >= 500 && retries > 0) {
          console.warn(`Retrying API request to ${endpoint}, ${retries} retries left`)
          return fetchApi(endpoint, options, retries - 1)
        }
        
        return {
          success: false,
          message: errorMessage,
        }
      }

      const data = await response.json()
      return data as ApiResponse<T>
    } catch (err) {
      // Handle different error types
      let errorType = ApiErrorType.UNKNOWN
      let errorMessage = 'An unknown error occurred'
      
      if (err instanceof Error) {
        errorMessage = err.message
        
        if (err.name === 'AbortError') {
          errorType = ApiErrorType.TIMEOUT
          errorMessage = 'Request timed out'
        } else if ('TypeError' === err.name) {
          errorType = ApiErrorType.NETWORK
          errorMessage = 'Network error occurred'
        }
      }
      
      const apiError: ApiError = {
        type: errorType,
        message: errorMessage
      }
      
      setError(apiError)
      
      // Retry on network errors if retries left
      if (errorType === ApiErrorType.NETWORK && retries > 0) {
        console.warn(`Retrying API request to ${endpoint}, ${retries} retries left`)
        return fetchApi(endpoint, options, retries - 1)
      }
      
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
    formData: FormData,
    retries = 2
  ): Promise<ApiResponse<T>> => {
    setLoading(true)
    setError(null)
    
    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 30000) // 30 second timeout for uploads
      
      const response = await fetch(`/api${endpoint}`, {
        method: 'POST',
        body: formData,
        signal: controller.signal,
      })
      
      clearTimeout(timeoutId)

      if (!response.ok) {
        // Handle different error status codes
        let errorType = ApiErrorType.SERVER
        let errorMessage = `API request failed with status ${response.status}`
        
        try {
          // Try to parse error message from response
          const errorData = await response.json()
          if (errorData && errorData.message) {
            errorMessage = errorData.message
          }
        } catch (e) {
          // Ignore JSON parse error and use default message
        }
        
        const apiError: ApiError = {
          type: errorType,
          message: errorMessage,
          status: response.status
        }
        
        setError(apiError)
        
        // Retry on server errors (5xx) if retries left
        if (response.status >= 500 && retries > 0) {
          console.warn(`Retrying form submission to ${endpoint}, ${retries} retries left`)
          return postFormData(endpoint, formData, retries - 1)
        }
        
        return {
          success: false,
          message: errorMessage,
        }
      }

      const data = await response.json()
      return data as ApiResponse<T>
    } catch (err) {
      // Handle different error types
      let errorType = ApiErrorType.UNKNOWN
      let errorMessage = 'An unknown error occurred'
      
      if (err instanceof Error) {
        errorMessage = err.message
        
        if (err.name === 'AbortError') {
          errorType = ApiErrorType.TIMEOUT
          errorMessage = 'Request timed out'
        } else if ('TypeError' === err.name) {
          errorType = ApiErrorType.NETWORK
          errorMessage = 'Network error occurred'
        }
      }
      
      const apiError: ApiError = {
        type: errorType,
        message: errorMessage
      }
      
      setError(apiError)
      
      // Retry on network errors if retries left
      if (errorType === ApiErrorType.NETWORK && retries > 0) {
        console.warn(`Retrying form submission to ${endpoint}, ${retries} retries left`)
        return postFormData(endpoint, formData, retries - 1)
      }
      
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