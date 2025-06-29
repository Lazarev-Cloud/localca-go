import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData()
    
    // Forward the request to the Go backend
    const response = await fetch(`${config.apiUrl}/api/delete`, {
      method: 'POST',
      body: formData,
    })

    if (!response.ok) {
      throw new Error(`Backend returned ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error deleting certificate:', error)
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Failed to delete certificate' 
      },
      { status: 500 }
    )
  }
} 