import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

// Direct login proxy to handle authentication specifically
export async function POST(request: NextRequest) {
  try {
    // Get form data or JSON
    const contentType = request.headers.get('content-type')
    let body
    
    if (contentType?.includes('application/json')) {
      body = JSON.stringify(await request.json())
    } else {
      body = await request.text()
    }

    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/login`, {
      method: 'POST',
      headers: {
        'Content-Type': contentType || 'application/x-www-form-urlencoded',
        'Cookie': cookieHeader,
      },
      credentials: 'include',
      body: body,
    })

    if (!response.ok) {
      const errorData = await response.json()
      return NextResponse.json(errorData, { status: response.status })
    }

    const data = await response.json()
    
    // Create response
    const nextResponse = NextResponse.json(data)
    
    // Forward Set-Cookie headers
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.append('Set-Cookie', value)
      }
    })

    return nextResponse
  } catch (error) {
    console.error('Error in login proxy:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Login failed' 
      },
      { status: 500 }
    )
  }
} 