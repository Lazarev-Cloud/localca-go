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
    console.log(`Proxying ${method} request to ${apiPath}`)

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
    
    // Add content-type if it's not already set
    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'application/json'
    }
    
    console.log('Proxy request headers:', headers)

    // Get request body for POST/PUT requests
    let body = undefined
    if (method === 'POST' || method === 'PUT') {
      if (request.headers.get('content-type')?.includes('application/json')) {
        body = JSON.stringify(await request.json())
      } else if (request.headers.get('content-type')?.includes('application/x-www-form-urlencoded')) {
        body = await request.text()
      } else {
        body = await request.text()
      }
    }

    // Add timeout for the backend request
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

    // Make the request to the backend
    const response = await fetch(`${config.apiUrl}/${apiPath}`, {
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

    // Create the NextResponse
    const nextResponse = NextResponse.json(responseData, {
      status: response.status,
    })

    // Forward all response headers
    response.headers.forEach((value, key) => {
      // Handle Set-Cookie specially to ensure it's properly forwarded
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.set('Set-Cookie', value)
      } else {
        nextResponse.headers.set(key, value)
      }
    })

    return nextResponse
  } catch (error) {
    console.error(`Proxy error for ${method} /${pathSegments.join('/')}:`, error)
    
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
      } else if (error.message.includes('ECONNREFUSED')) {
        return NextResponse.json(
          { 
            success: false, 
            message: 'Could not connect to backend service',
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