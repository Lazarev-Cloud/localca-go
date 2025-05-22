import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function GET(request: NextRequest) {
  try {
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')
    
    console.log('Debug route - Cookies from request:', cookies)
    console.log('Debug route - Cookie header being sent to backend:', cookieHeader)

    // Add timeout for the backend request
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

    // Make a direct request to the Go backend
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

    console.log('Debug route - Backend response status:', response.status)
    console.log('Debug route - Backend response headers:', Object.fromEntries([...response.headers.entries()]))
    
    let responseBody;
    try {
      responseBody = await response.json();
      console.log('Debug route - Backend response body:', responseBody);
    } catch (err) {
      console.error('Debug route - Error parsing response body:', err);
      responseBody = { error: 'Failed to parse response body' };
    }

    // Return all debugging information
    return NextResponse.json({
      request: {
        cookies: cookies,
        cookieHeader: cookieHeader
      },
      response: {
        status: response.status,
        headers: Object.fromEntries([...response.headers.entries()]),
        body: responseBody
      }
    })
  } catch (error) {
    console.error('Debug route - Error:', error)
    
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Debug route error',
        error: String(error)
      },
      { status: 500 }
    )
  }
} 