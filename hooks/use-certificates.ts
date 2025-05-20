"use client"

import { useState, useEffect } from 'react'
import { useApi } from './use-api'

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
  const { loading, error, fetchApi, postFormData } = useApi()

  const fetchCertificates = async () => {
    const response = await fetchApi<{ certificates: Certificate[] }>('/certificates')
    if (response.success && response.data) {
      setCertificates(response.data.certificates || [])
    }
  }

  const fetchCAInfo = async () => {
    const response = await fetchApi<CAInfo>('/ca-info')
    if (response.success && response.data) {
      setCAInfo(response.data)
    }
  }

  const createCertificate = async (formData: FormData) => {
    const response = await postFormData<any>('/certificates', formData)
    if (response.success) {
      // Refresh the certificates list
      fetchCertificates()
    }
    return response
  }

  const revokeCertificate = async (serialNumber: string) => {
    const formData = new FormData()
    formData.append('serial_number', serialNumber)
    
    const response = await postFormData<any>('/revoke', formData)
    if (response.success) {
      // Refresh the certificates list
      fetchCertificates()
    }
    return response
  }

  const renewCertificate = async (serialNumber: string) => {
    const formData = new FormData()
    formData.append('serial_number', serialNumber)
    
    const response = await postFormData<any>('/renew', formData)
    if (response.success) {
      // Refresh the certificates list
      fetchCertificates()
    }
    return response
  }

  const deleteCertificate = async (serialNumber: string) => {
    const formData = new FormData()
    formData.append('serial_number', serialNumber)
    
    const response = await postFormData<any>('/delete', formData)
    if (response.success) {
      // Refresh the certificates list
      fetchCertificates()
    }
    return response
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