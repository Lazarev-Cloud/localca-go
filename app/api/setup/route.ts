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
export async function GET() {
  try {
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/setup`, {
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const data = await response.json()

    // Return the response from the backend
    return NextResponse.json(data, {
      status: response.status,
    })
  } catch (error) {
    console.error('Error in setup GET:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to get setup info' 
      },
      { status: 500 }
    )
  }
}

// POST handler for setup
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/setup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()

    // Return the response from the backend
    return NextResponse.json(data, {
      status: response.status,
    })
  } catch (error) {
    console.error('Error in setup:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to complete setup' 
      },
      { status: 500 }
    )
  }
} 