import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

// GET handler to fetch all certificates
export async function GET(request: NextRequest) {
  try {
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')

    // Make a request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/certificates`, {
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookieHeader,
      },
      credentials: 'include',
      cache: 'no-store',
    })

    // Check for unauthorized response which likely means setup is required
    if (response.status === 401) {
      // Return a specific response that the frontend can use to redirect
      return NextResponse.json(
        { 
          success: false, 
          message: 'Setup required',
          setupRequired: true
        },
        { status: 401 }
      )
    }

    if (!response.ok) {
      throw new Error(`Backend returned ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching certificates:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to fetch certificates' 
      },
      { status: 500 }
    )
  }
}

// POST handler to create a new certificate
export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData()
    
    // Get cookies from the request
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')
    
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/certificates`, {
      method: 'POST',
      headers: {
        'Cookie': cookieHeader,
      },
      credentials: 'include',
      body: formData,
    })

    // Check for unauthorized response which likely means setup is required
    if (response.status === 401) {
      // Return a specific response that the frontend can use to redirect
      return NextResponse.json(
        { 
          success: false, 
          message: 'Setup required',
          setupRequired: true
        },
        { status: 401 }
      )
    }

    if (!response.ok) {
      throw new Error(`Backend returned ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error creating certificate:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to create certificate' 
      },
      { status: 500 }
    )
  }
} 