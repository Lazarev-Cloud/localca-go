import { NextRequest } from 'next/server'

// Get cookie header string from request
export function getCookieHeader(request: NextRequest): string {
  const cookies = request.cookies.getAll()
  return cookies
    .map(c => `${c.name}=${c.value}`)
    .join('; ')
}

// Safely get cookies from a fetch response
export function getSetCookieHeader(response: Response): string | null {
  return response.headers.get('set-cookie')
} 