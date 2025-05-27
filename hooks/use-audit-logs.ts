"use client"

import { useState, useEffect, useCallback, useRef } from 'react'
import { useApi } from './use-api'

interface AuditLog {
  id: number
  action: string
  resource: string
  resource_id?: string
  user_ip?: string
  user_agent?: string
  details?: string
  success: boolean
  error?: string
  created_at: string
}

interface AuditLogsResponse {
  audit_logs: AuditLog[]
  total: number
  limit: number
  offset: number
}

// Global cache to prevent duplicate requests
let globalCache: {
  data: AuditLog[] | null
  timestamp: number
  loading: boolean
} = {
  data: null,
  timestamp: 0,
  loading: false
}

const CACHE_DURATION = 10000 // 10 seconds
const MIN_REQUEST_INTERVAL = 2000 // 2 seconds between requests

export function useAuditLogs(limit: number = 10, offset: number = 0) {
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { fetchApi } = useApi()
  const lastRequestRef = useRef<number>(0)

  const fetchAuditLogs = useCallback(async (force: boolean = false) => {
    const now = Date.now()
    
    // Check if we should use cached data
    if (!force && globalCache.data && (now - globalCache.timestamp) < CACHE_DURATION) {
      setAuditLogs(globalCache.data.slice(offset, offset + limit))
      setLoading(false)
      return
    }

    // Prevent too frequent requests
    if (!force && (now - lastRequestRef.current) < MIN_REQUEST_INTERVAL) {
      return
    }

    // Prevent duplicate requests
    if (globalCache.loading) {
      return
    }

    try {
      globalCache.loading = true
      setLoading(true)
      setError(null)
      lastRequestRef.current = now

      const response = await fetchApi<AuditLogsResponse>(`/api/audit-logs?limit=50&offset=0`)
      
      if (response.success && response.data) {
        globalCache.data = response.data.audit_logs || []
        globalCache.timestamp = now
        setAuditLogs(globalCache.data.slice(offset, offset + limit))
      } else {
        setAuditLogs([])
        setError('Failed to load audit logs')
      }
    } catch (err) {
      console.error('Failed to fetch audit logs:', err)
      setError('Failed to load audit logs')
      setAuditLogs([])
    } finally {
      globalCache.loading = false
      setLoading(false)
    }
  }, [fetchApi, limit, offset])

  useEffect(() => {
    fetchAuditLogs()
  }, []) // Only run once on mount

  const refresh = useCallback(() => {
    fetchAuditLogs(true)
  }, [fetchAuditLogs])

  return {
    auditLogs,
    loading,
    error,
    refresh
  }
} 