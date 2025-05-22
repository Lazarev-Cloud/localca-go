import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
      },
      credentials: 'include',
    })

    const data = await response.json()
    
    // Create response
    const nextResponse = NextResponse.json(data, { status: response.status })
    
    // Forward Set-Cookie headers (which should clear the session cookie)
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.append('Set-Cookie', value)
      }
    })

    return nextResponse
  } catch (error) {
    console.error('Error in logout proxy:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Logout failed' 
      },
      { status: 500 }
    )
  }
} 