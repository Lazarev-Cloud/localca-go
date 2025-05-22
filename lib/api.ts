/**
 * LocalCA API Client
 *
 * This module provides functions to interact with the LocalCA backend API.
 */

// Base URL from environment variable
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ""

// Error handling helper
class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
    this.name = "ApiError"
  }
}

// Helper function to handle API responses
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorText = await response.text()
    let errorMessage

    try {
      const errorJson = JSON.parse(errorText)
      errorMessage = errorJson.message || errorJson.error || `API error: ${response.status}`
    } catch (e) {
      errorMessage = errorText || `API error: ${response.status}`
    }

    throw new ApiError(errorMessage, response.status)
  }

  return await response.json()
}

// Get CSRF token from cookie or meta tag
function getCSRFToken(): string {
  // Check for CSRF token in meta tag
  const metaTag = document.querySelector('meta[name="csrf-token"]')
  if (metaTag) {
    return metaTag.getAttribute("content") || ""
  }

  // Fallback to cookie
  const cookies = document.cookie.split(";")
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split("=")
    if (name === "csrf_token") {
      return value
    }
  }

  return ""
}

// Types for API responses
export interface Certificate {
  id: string
  common_name: string
  type: "Server" | "Client"
  expiry_date: string
  issued_date: string
  is_expiring_soon: boolean
  is_expired: boolean
  is_revoked: boolean
  serial_number: string
  organization?: string
  country?: string
  alternative_names?: string[]
  key_usage?: string[]
  extended_key_usage?: string[]
  key_type?: string
  key_size?: string
  signature_algorithm?: string
  fingerprint?: string
}

export interface CAInfo {
  common_name: string
  organization: string
  country: string
  valid_until: string
  status: "Active" | "Inactive"
  remaining_days: number
}

export interface SystemStatus {
  storage_usage: number
  database_size: number
  uptime: number
  certificate_count: number
  expiring_soon_count: number
  revoked_count: number
  client_certificate_count: number
  alerts: Array<{
    type: string
    message: string
  }>
}

export interface RecentActivity {
  id: string
  type: "create" | "download" | "renew" | "revoke"
  message: string
  timestamp: string
}

export interface Settings {
  ca_name: string
  organization: string
  country: string
  tls_enabled: boolean
  email_notify: boolean
  smtp_server: string
  smtp_port: string
  smtp_user: string
  smtp_password: string
  smtp_use_tls: boolean
  email_from: string
  email_to: string
  storage_path: string
  backup_path: string
  auto_backup: boolean
  crl_expiry_days: string
}

export interface ApiResponse<T = any> {
  success: boolean
  message?: string
  data?: T
}

// API Functions

/**
 * Get information about the Certificate Authority
 */
export async function getCAInfo(): Promise<CAInfo> {
  const response = await fetch(`${API_BASE_URL}/ca-info`)
  return handleResponse<CAInfo>(response)
}

/**
 * Get system status information
 */
export async function getSystemStatus(): Promise<SystemStatus> {
  const response = await fetch(`${API_BASE_URL}/system-status`)
  return handleResponse<SystemStatus>(response)
}

/**
 * Get list of all certificates
 */
export async function getCertificates(): Promise<Certificate[]> {
  const response = await fetch(`${API_BASE_URL}/certificates`)
  return handleResponse<Certificate[]>(response)
}

/**
 * Get details of a specific certificate
 */
export async function getCertificate(id: string): Promise<Certificate> {
  const response = await fetch(`${API_BASE_URL}/certificates/${id}`)
  return handleResponse<Certificate>(response)
}

/**
 * Create a new certificate
 */
export async function createCertificate(certificateData: any): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/certificates`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfToken,
    },
    body: JSON.stringify(certificateData),
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}

/**
 * Revoke a certificate
 */
export async function revokeCertificate(id: string): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/certificates/${id}/revoke`, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrfToken,
    },
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}

/**
 * Renew a certificate
 */
export async function renewCertificate(id: string): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/certificates/${id}/renew`, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrfToken,
    },
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}

/**
 * Download a certificate
 * Returns the URL to download the certificate
 */
export function getCertificateDownloadUrl(id: string, format: "pem" | "p12" | "key" = "pem"): string {
  return `${API_BASE_URL}/certificates/${id}/download?format=${format}`
}

/**
 * Download the CA certificate
 * Returns the URL to download the CA certificate
 */
export function getCADownloadUrl(format: "pem" | "chain" | "crl" = "pem"): string {
  return `${API_BASE_URL}/ca/download?format=${format}`
}

/**
 * Get recent activity
 */
export async function getRecentActivity(): Promise<RecentActivity[]> {
  const response = await fetch(`${API_BASE_URL}/activity`)
  return handleResponse<RecentActivity[]>(response)
}

/**
 * Get system settings
 */
export async function getSettings(): Promise<Settings> {
  const response = await fetch(`${API_BASE_URL}/settings`)
  return handleResponse<Settings>(response)
}

/**
 * Update system settings
 */
export async function updateSettings(settings: Partial<Settings>): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/settings`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfToken,
    },
    body: JSON.stringify(settings),
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}

/**
 * Send test email
 */
export async function sendTestEmail(email: string): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/test-email`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfToken,
    },
    body: JSON.stringify({ email }),
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}

/**
 * Regenerate CRL
 */
export async function regenerateCRL(): Promise<ApiResponse> {
  const csrfToken = getCSRFToken()

  const response = await fetch(`${API_BASE_URL}/ca/regenerate-crl`, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrfToken,
    },
    credentials: "include",
  })

  return handleResponse<ApiResponse>(response)
}
