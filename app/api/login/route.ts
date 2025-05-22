import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
      credentials: 'include',
    })

    // Get cookies from backend response
    const setCookieHeader = response.headers.get('set-cookie')
    
    // Create the response object from the backend data
    const data = await response.json()
    const nextResponse = NextResponse.json(data, {
      status: response.status,
    })

    // Forward cookies from backend to client if available
    if (setCookieHeader) {
      // Parse the cookie to set it properly
      const cookieParts = setCookieHeader.split(';')
      const mainPart = cookieParts[0].split('=')
      
      if (mainPart.length === 2) {
        const cookieName = mainPart[0].trim()
        const cookieValue = mainPart[1].trim()
        
        // Determine if cookie should be secure
        const isSecure = setCookieHeader.toLowerCase().includes('secure')
        
        // Set cookie with appropriate attributes - ensure path is / and samesite is lax
        nextResponse.cookies.set({
          name: cookieName,
          value: cookieValue,
          path: '/',
          httpOnly: true,
          secure: isSecure,
          sameSite: 'lax',
        })
      } else {
        // Fallback to original method if parsing fails
        nextResponse.headers.set('set-cookie', setCookieHeader)
      }
    }

    return nextResponse
  } catch (error) {
    console.error('Login error:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Login failed' 
      },
      { status: 500 }
    )
  }
} 