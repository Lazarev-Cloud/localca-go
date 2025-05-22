import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

// Direct login proxy to handle authentication specifically
export async function POST(request: NextRequest) {
  const backendUrl = `${config.apiUrl}/api/login`
  console.log(`Direct login proxy: POSTing to ${backendUrl}`)
  
  try {
    // Get request body
    const body = await request.text()
    
    // Forward cookies
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map((c: any) => `${c.name}=${c.value}`)
      .join('; ')
    
    // Prepare headers
    const headers: Record<string, string> = {}
    request.headers.forEach((value, key) => {
      if (key.toLowerCase() !== 'host') {
        headers[key] = value
      }
    })
    
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader
    }
    
    // Add timeout
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000)
    
    // Send request to backend
    const response = await fetch(backendUrl, {
      method: 'POST',
      headers,
      body,
      credentials: 'include',
      cache: 'no-store',
      signal: controller.signal,
    })
    
    clearTimeout(timeoutId)
    
    // Handle response
    let responseData
    const contentType = response.headers.get('content-type')
    
    if (contentType?.includes('application/json')) {
      responseData = await response.json()
      console.log(`Login response (${response.status}):`, responseData)
    } else {
      responseData = await response.text()
      console.log(`Login text response (${response.status}): ${responseData.substring(0, 100)}...`)
    }
    
    // Create response - handle both JSON and non-JSON responses
    let nextResponse
    if (contentType?.includes('application/json')) {
      nextResponse = NextResponse.json(responseData, {
        status: response.status,
      })
    } else {
      // For non-JSON responses, return as text but with JSON wrapper for consistency
      nextResponse = NextResponse.json({
        success: response.ok,
        message: responseData
      }, {
        status: response.status,
      })
    }
    
    // Forward response headers, handling Set-Cookie specially
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() === 'set-cookie') {
        // For Set-Cookie, we need to append rather than set to handle multiple cookies
        nextResponse.headers.append('Set-Cookie', value)
      } else if (key.toLowerCase() !== 'content-length' && key.toLowerCase() !== 'transfer-encoding') {
        // Skip headers that Next.js manages automatically
        nextResponse.headers.set(key, value)
      }
    })
    
    return nextResponse
  } catch (error) {
    console.error(`Login proxy error connecting to ${backendUrl}:`, error)
    
    return NextResponse.json(
      { 
        success: false,
        message: `Login failed: Could not connect to authentication service at ${config.apiUrl}`,
        error: error instanceof Error ? error.message : String(error)
      },
      { status: 503 }
    )
  }
} 