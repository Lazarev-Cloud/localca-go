import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData()
    const serialNumber = formData.get('serial_number')
    
    if (!serialNumber) {
      return NextResponse.json(
        { 
          success: false, 
          message: 'Serial number is required' 
        },
        { status: 400 }
      )
    }
    
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/revoke`, {
      method: 'POST',
      body: formData,
      headers: {
        // Don't set Content-Type here as it will be set automatically for FormData
      },
    })

    // Handle backend errors with specific status codes
    if (!response.ok) {
      let errorMessage = `Backend returned ${response.status}`
      
      try {
        const errorData = await response.json()
        if (errorData && errorData.message) {
          errorMessage = errorData.message
        }
      } catch (e) {
        // If we can't parse the error response, use the default message
      }
      
      // Pass through the status code from the backend
      return NextResponse.json(
        { 
          success: false, 
          message: errorMessage 
        },
        { status: response.status }
      )
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error revoking certificate:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to revoke certificate' 
      },
      { status: 500 }
    )
  }
} 