import { NextRequest, NextResponse } from 'next/server'

// GET handler to fetch all certificates
export async function GET(request: NextRequest) {
  try {
    // Make a request to the Go backend
    const response = await fetch('http://localhost:8080/api/certificates', {
      headers: {
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

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
    
    // Forward the request to the Go backend
    const response = await fetch('http://localhost:8080/api/certificates', {
      method: 'POST',
      body: formData,
    })

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