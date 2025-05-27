"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Download, RefreshCw, XCircle, MoreHorizontal, Loader2 } from "lucide-react"
import { useCertificates } from "@/hooks/use-certificates"
import { useToast } from "@/hooks/use-toast-new"
import { useRouter } from "next/navigation"

interface CertificateActionsProps {
  id: string
}

export function CertificateActions({ id }: CertificateActionsProps) {
  const { certificates, revokeCertificate, renewCertificate, deleteCertificate } = useCertificates()
  const { toast } = useToast()
  const router = useRouter()
  const [certificate, setCertificate] = useState<any>(null)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    // Find the certificate by serial number
    const cert = certificates.find(c => c.serial_number === id)
    setCertificate(cert)
  }, [certificates, id])

  const handleDownload = async (format: string = 'pem') => {
    if (!certificate) return
    
    try {
      setActionLoading(`download-${format}`)
      // Use the proxy endpoint to download certificate
      const response = await fetch(`/api/proxy/api/download/${encodeURIComponent(certificate.common_name)}/${format}`, {
        credentials: 'include'
      })
      
      if (!response.ok) {
        throw new Error('Failed to download certificate')
      }
      
      // Create download link
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${certificate.common_name}.${format}`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      toast({
        title: "Download successful",
        description: `Certificate downloaded as ${format.toUpperCase()}.`,
      })
    } catch (error) {
      toast({
        title: "Download failed",
        description: error instanceof Error ? error.message : "Failed to download certificate",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  const handleRenew = async () => {
    if (!certificate) return
    
    try {
      setActionLoading('renew')
      const result = await renewCertificate(certificate.serial_number)
      
      if (result.success) {
        toast({
          title: "Certificate renewed",
          description: `Certificate ${certificate.common_name} was renewed successfully.`,
        })
      } else {
        toast({
          title: "Renewal failed",
          description: result.message || "Failed to renew certificate",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Renewal failed",
        description: error instanceof Error ? error.message : "Failed to renew certificate",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  const handleRevoke = async () => {
    if (!certificate) return
    
    if (!confirm(`Are you sure you want to revoke the certificate "${certificate.common_name}"? This action cannot be undone.`)) {
      return
    }
    
    try {
      setActionLoading('revoke')
      const result = await revokeCertificate(certificate.serial_number)
      
      if (result.success) {
        toast({
          title: "Certificate revoked",
          description: `Certificate ${certificate.common_name} was revoked successfully.`,
        })
        // Redirect back to certificates list
        router.push('/certificates')
      } else {
        toast({
          title: "Revocation failed",
          description: result.message || "Failed to revoke certificate",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Revocation failed",
        description: error instanceof Error ? error.message : "Failed to revoke certificate",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  const handleDelete = async () => {
    if (!certificate) return
    
    if (!confirm(`Are you sure you want to delete the certificate "${certificate.common_name}"? This action cannot be undone.`)) {
      return
    }
    
    try {
      setActionLoading('delete')
      const result = await deleteCertificate(certificate.serial_number)
      
      if (result.success) {
        toast({
          title: "Certificate deleted",
          description: `Certificate ${certificate.common_name} was deleted successfully.`,
        })
        // Redirect back to certificates list
        router.push('/certificates')
      } else {
        toast({
          title: "Deletion failed",
          description: result.message || "Failed to delete certificate",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Deletion failed",
        description: error instanceof Error ? error.message : "Failed to delete certificate",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  if (!certificate) {
    return null
  }

  return (
    <div className="flex items-center gap-2">
      <Button 
        variant="outline" 
        className="flex items-center gap-2"
        onClick={() => handleDownload('pem')}
        disabled={actionLoading === 'download-pem'}
      >
        {actionLoading === 'download-pem' ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Download className="h-4 w-4" />
        )}
        Download
      </Button>
      {!certificate.is_revoked && (
        <Button 
          variant="outline" 
          className="flex items-center gap-2"
          onClick={handleRenew}
          disabled={actionLoading === 'renew'}
        >
          {actionLoading === 'renew' ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="h-4 w-4" />
          )}
          Renew
        </Button>
      )}
      {!certificate.is_revoked && (
        <Button 
          variant="outline" 
          className="flex items-center gap-2 text-red-500 hover:text-red-600"
          onClick={handleRevoke}
          disabled={actionLoading === 'revoke'}
        >
          {actionLoading === 'revoke' ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <XCircle className="h-4 w-4" />
          )}
          Revoke
        </Button>
      )}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>More Actions</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem 
            onClick={() => handleDownload('pem')}
            disabled={actionLoading === 'download-pem'}
          >
            {actionLoading === 'download-pem' ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Download className="mr-2 h-4 w-4" />
            )}
            Download as PEM
          </DropdownMenuItem>
          {certificate.is_client && (
            <DropdownMenuItem 
              onClick={() => handleDownload('p12')}
              disabled={actionLoading === 'download-p12'}
            >
              {actionLoading === 'download-p12' ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Download className="mr-2 h-4 w-4" />
              )}
              Download as P12
            </DropdownMenuItem>
          )}
          <DropdownMenuSeparator />
          <DropdownMenuItem 
            onClick={handleDelete}
            disabled={actionLoading === 'delete'}
            className="text-red-600"
          >
            {actionLoading === 'delete' ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <XCircle className="mr-2 h-4 w-4" />
            )}
            Delete Certificate
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
