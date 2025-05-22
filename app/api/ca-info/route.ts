import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function GET(request: NextRequest) {
  try {
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Add timeout for the backend request
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

    // Make a request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/ca-info`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader, // Forward cookies to backend
      },
      credentials: 'include',
      cache: 'no-store',
      signal: controller.signal,
    })
    
    // Clear timeout
    clearTimeout(timeoutId)

    // Check if setup is required
    if (response.status === 401) {
      try {
        // Attempt to parse the response to determine if it's a setup or auth issue
        const data = await response.json()
        
        // Check if this is a setup required message - handle both possible formats
        if ((data && data.data && data.data.setup_required === true) ||
            (data && data.setupRequired === true) ||
            (data && data.message === 'Setup required')) {
          return NextResponse.json(
            { 
              success: false, 
              message: 'Setup required',
              setupRequired: true
            },
            { status: 401 }
          )
        }
        
        // Otherwise it's just a regular auth issue
        return NextResponse.json(
          { 
            success: false, 
            message: 'Authentication required',
            setupRequired: false
          },
          { status: 401 }
        )
      } catch (err) {
        // If we can't parse the response, default to auth required
        return NextResponse.json(
          { 
            success: false, 
            message: 'Authentication required',
            setupRequired: false
          },
          { status: 401 }
        )
      }
    }

    if (!response.ok) {
      throw new Error(`Backend returned ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching CA info:', error)
    
    // Check if the error is a timeout or connection error
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        return NextResponse.json(
          { 
            success: false, 
            message: 'Connection to backend timed out', 
            retryable: true
          },
          { status: 503 }
        )
      } else if (error.message.includes('ECONNREFUSED')) {
        return NextResponse.json(
          { 
            success: false, 
            message: 'Could not connect to backend service',
            retryable: true
          },
          { status: 503 }
        )
      }
    }
    
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to fetch CA info' 
      },
      { status: 500 }
    )
  }
} 