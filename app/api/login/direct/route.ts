import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Create form data for the Go backend (it expects form data, not JSON)
    const formData = new URLSearchParams()
    formData.append('username', body.username)
    formData.append('password', body.password)
    
    console.log('Login attempt for username:', body.username)
    
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
    console.log('Set-Cookie from backend:', setCookieHeader)

    // Create success response
    const nextResponse = NextResponse.json({
      success: true,
      message: 'Login successful',
    })

    // Forward cookies from backend to client
    if (setCookieHeader) {
      // Simply forward the entire Set-Cookie header to client
      // This is the most reliable approach
      nextResponse.headers.set('Set-Cookie', setCookieHeader)
      
      // Log what we're sending back
      console.log('Forwarding Set-Cookie header to client:', setCookieHeader)
      
      // For debugging, also try to parse individual cookies
      try {
        const cookies = setCookieHeader.split(/,(?=\s[A-Za-z0-9]+=)/);
        console.log('Parsed individual cookies:', cookies);
      } catch (err) {
        console.error('Error parsing cookies:', err);
      }
    } else {
      console.log('No cookies received from backend');
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