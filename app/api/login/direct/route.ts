import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Create form data for the Go backend (it expects form data, not JSON)
    const formData = new URLSearchParams()
    formData.append('username', body.username)
    formData.append('password', body.password)
    
    // Forward the request to the Go backend's regular login endpoint
    // Use full URL with timeout to handle potential network issues
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout
    
    const response = await fetch(`${config.apiUrl}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData.toString(),
      redirect: 'manual', // Don't follow redirects
      signal: controller.signal,
    })
    
    clearTimeout(timeoutId)

    // Get cookies from backend response
    const setCookieHeader = response.headers.get('set-cookie')

    // Create success response
    const nextResponse = NextResponse.json({
      success: true,
      message: 'Login successful',
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
        
        // Set cookie with appropriate attributes
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