"use client"

import { useState, useEffect } from 'react'

export interface Certificate {
  common_name: string
  expiry_date: string
  is_client: boolean
  serial_number: string
  is_expired: boolean
  is_expiring_soon: boolean
  is_revoked: boolean
}

export interface CAInfo {
  common_name: string
  organization: string
  country: string
  expiry_date: string
  is_expired: boolean
}

export function useCertificates() {
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const [caInfo, setCAInfo] = useState<CAInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchCertificates = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch('/api/certificates')
      if (!response.ok) {
        throw new Error(`Failed to fetch certificates: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        setCertificates(data.data.certificates || [])
      } else {
        throw new Error(data.message || 'Failed to fetch certificates')
      }
    } catch (err) {
      console.error('Error fetching certificates:', err)
      setError(err instanceof Error ? err.message : 'An unknown error occurred')
    } finally {
      setLoading(false)
    }
  }

  const fetchCAInfo = async () => {
    try {
      const response = await fetch('/api/ca-info')
      if (!response.ok) {
        throw new Error(`Failed to fetch CA info: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        setCAInfo(data.data)
      } else {
        throw new Error(data.message || 'Failed to fetch CA info')
      }
    } catch (err) {
      console.error('Error fetching CA info:', err)
      // We don't set the error state here to avoid blocking the UI if only CA info fails
    }
  }

  const createCertificate = async (formData: FormData) => {
    try {
      const response = await fetch('/api/certificates', {
        method: 'POST',
        body: formData,
      })
      
      if (!response.ok) {
        throw new Error(`Failed to create certificate: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        // Refresh the certificates list
        fetchCertificates()
        return { success: true, data: data.data }
      } else {
        throw new Error(data.message || 'Failed to create certificate')
      }
    } catch (err) {
      console.error('Error creating certificate:', err)
      return { 
        success: false, 
        error: err instanceof Error ? err.message : 'An unknown error occurred' 
      }
    }
  }

  const revokeCertificate = async (serialNumber: string) => {
    try {
      const formData = new FormData()
      formData.append('serial_number', serialNumber)
      
      const response = await fetch('/api/revoke', {
        method: 'POST',
        body: formData,
      })
      
      if (!response.ok) {
        throw new Error(`Failed to revoke certificate: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        // Refresh the certificates list
        fetchCertificates()
        return { success: true }
      } else {
        throw new Error(data.message || 'Failed to revoke certificate')
      }
    } catch (err) {
      console.error('Error revoking certificate:', err)
      return { 
        success: false, 
        error: err instanceof Error ? err.message : 'An unknown error occurred' 
      }
    }
  }

  const renewCertificate = async (serialNumber: string) => {
    try {
      const formData = new FormData()
      formData.append('serial_number', serialNumber)
      
      const response = await fetch('/api/renew', {
        method: 'POST',
        body: formData,
      })
      
      if (!response.ok) {
        throw new Error(`Failed to renew certificate: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        // Refresh the certificates list
        fetchCertificates()
        return { success: true, data: data.data }
      } else {
        throw new Error(data.message || 'Failed to renew certificate')
      }
    } catch (err) {
      console.error('Error renewing certificate:', err)
      return { 
        success: false, 
        error: err instanceof Error ? err.message : 'An unknown error occurred' 
      }
    }
  }

  const deleteCertificate = async (serialNumber: string) => {
    try {
      const formData = new FormData()
      formData.append('serial_number', serialNumber)
      
      const response = await fetch('/api/delete', {
        method: 'POST',
        body: formData,
      })
      
      if (!response.ok) {
        throw new Error(`Failed to delete certificate: ${response.status}`)
      }
      
      const data = await response.json()
      if (data.success) {
        // Refresh the certificates list
        fetchCertificates()
        return { success: true }
      } else {
        throw new Error(data.message || 'Failed to delete certificate')
      }
    } catch (err) {
      console.error('Error deleting certificate:', err)
      return { 
        success: false, 
        error: err instanceof Error ? err.message : 'An unknown error occurred' 
      }
    }
  }

  useEffect(() => {
    fetchCertificates()
    fetchCAInfo()
  }, [])

  return {
    certificates,
    caInfo,
    loading,
    error,
    fetchCertificates,
    fetchCAInfo,
    createCertificate,
    revokeCertificate,
    renewCertificate,
    deleteCertificate
  }
} 