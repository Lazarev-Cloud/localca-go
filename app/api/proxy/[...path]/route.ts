import { NextRequest, NextResponse } from 'next/server'
import config from '@/lib/config'

export async function GET(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return await proxyRequest(request, params.path, 'GET')
}

export async function POST(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return await proxyRequest(request, params.path, 'POST')
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return await proxyRequest(request, params.path, 'PUT')
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return await proxyRequest(request, params.path, 'DELETE')
}

async function proxyRequest(
  request: NextRequest,
  pathSegments: string[],
  method: string
) {
  try {
    // Construct the path for the backend API
    const apiPath = pathSegments.join('/')
    
    // Get and forward all cookies
    const cookies = request.cookies.getAll()
    const cookieHeader = cookies
      .map(c => `${c.name}=${c.value}`)
      .join('; ')
    
    // Get and forward all headers (except host)
    const headers: Record<string, string> = {}
    request.headers.forEach((value, key) => {
      if (key.toLowerCase() !== 'host') {
        headers[key] = value
      }
    })
    
    // Add cookie header if we have cookies
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader
    }
    
    // Add CSRF token header if it exists in cookies
    const csrfCookie = cookies.find(c => c.name === 'csrf_token')
    if (csrfCookie) {
      headers['X-CSRF-Token'] = csrfCookie.value
    }
    
    // Get request body for POST/PUT requests without modifying content-type
    let body = undefined
    if (method === 'POST' || method === 'PUT') {
      const contentType = request.headers.get('content-type')
      if (contentType?.includes('application/json')) {
        // For JSON, parse and re-stringify to ensure valid JSON
        try {
          const jsonData = await request.json()
          body = JSON.stringify(jsonData)
        } catch (error) {
          console.error('Invalid JSON in request:', error)
          body = await request.text()
        }
      } else {
        // For all other content types (form data, text, etc.), pass as-is
        body = await request.text()
      }
    }

    // Add timeout for the backend request
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 10000) // 10 second timeout

    // Make the request to the backend
    // Ensure we're using the configured API URL, not a hardcoded localhost
    const backendUrl = config.apiUrl 
      ? `${config.apiUrl}/${apiPath}` // Use configured API URL if available
      : `/${apiPath}`; // Otherwise use relative URL
    
    const response = await fetch(backendUrl, {
      method,
      headers,
      body,
      credentials: 'include',
      cache: 'no-store',
      signal: controller.signal,
    })
    
    // Clear timeout
    clearTimeout(timeoutId)

    // Read response data
    const contentType = response.headers.get('content-type')
    let responseData: any
    
    if (contentType?.includes('application/json')) {
      responseData = await response.json()
    } else {
      responseData = await response.text()
    }

    // Create the NextResponse - handle both JSON and non-JSON responses
    let nextResponse
    if (contentType?.includes('application/json')) {
      nextResponse = NextResponse.json(responseData, {
        status: response.status,
      })
    } else {
      nextResponse = new NextResponse(responseData, {
        status: response.status,
        headers: {
          'Content-Type': contentType || 'text/plain'
        }
      })
    }

    // Forward all response headers
    response.headers.forEach((value, key) => {
      // Handle Set-Cookie specially to ensure it's properly forwarded
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.append('Set-Cookie', value)
      } else if (key.toLowerCase() !== 'content-length' && key.toLowerCase() !== 'transfer-encoding') {
        // Skip headers that Next.js manages automatically
        nextResponse.headers.set(key, value)
      }
    })

    return nextResponse
  } catch (error) {
    // Check if the error is a timeout or connection error
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        return NextResponse.json(
          { 
            success: false, 
            message: 'Connection to backend timed out', 
            retryable: true
          },
          { status: 503 }
        )
      } else if (
        error.message.includes('ECONNREFUSED') || 
        (error as any).code === 'ECONNREFUSED'
      ) {
        return NextResponse.json(
          { 
            success: false, 
            message: `Could not connect to backend service. Please ensure the Go server is running.`,
            retryable: true
          },
          { status: 503 }
        )
      }
    }
    
    return NextResponse.json(
      { 
        success: false, 
        message: error instanceof Error ? error.message : 'Proxy request failed' 
      },
      { status: 500 }
    )
  }
} 