import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

// Handle preflight OPTIONS request
export async function OPTIONS() {
  return new NextResponse(null, {
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    },
  })
}

// GET handler for setup
export async function GET(request: NextRequest) {
  try {
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/setup`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
      },
      credentials: 'include',
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Error in setup GET proxy:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Setup request failed' 
      },
      { status: 500 }
    )
  }
}

// POST handler for setup
export async function POST(request: NextRequest) {
  try {
    // Get JSON data
    const body = JSON.stringify(await request.json())
    
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/setup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
      },
      credentials: 'include',
      body: body,
    })

    const data = await response.json()
    
    // Create response
    const nextResponse = NextResponse.json(data, { status: response.status })
    
    // Forward Set-Cookie headers
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.append('Set-Cookie', value)
      }
    })

    return nextResponse
  } catch (error) {
    console.error('Error in setup POST proxy:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Setup failed' 
      },
      { status: 500 }
    )
  }
} 