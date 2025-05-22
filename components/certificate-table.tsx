"use client"

import { useState, useMemo } from "react"
import Link from "next/link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { MoreHorizontal, Download, RefreshCw, XCircle, AlertTriangle, CheckCircle, FileText, Loader2 } from "lucide-react"
import { useCertificates } from "@/hooks/use-certificates"
import { useToast } from "@/hooks/use-toast-new"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { useCertificateFilters } from "./certificate-filters"

export function CertificateTable() {
  const { certificates, loading, error, revokeCertificate, renewCertificate, deleteCertificate } = useCertificates()
  const { filters } = useCertificateFilters()
  const { toast } = useToast()
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  // Filter certificates based on the current filters
  const filteredCertificates = useMemo(() => {
    return certificates.filter(cert => {
      // Search filter
      if (filters.searchQuery) {
        const query = filters.searchQuery.toLowerCase()
        const matchesCommonName = cert.common_name.toLowerCase().includes(query)
        const matchesSerial = cert.serial_number.toLowerCase().includes(query)
        if (!matchesCommonName && !matchesSerial) {
          return false
        }
      }

      // Type filter
      if (filters.certificateType !== "all") {
        const isClient = filters.certificateType === "client"
        if (cert.is_client !== isClient) {
          return false
        }
      }

      // Status filter
      if (filters.status !== "all") {
        switch (filters.status) {
          case "valid":
            if (cert.is_revoked || cert.is_expired || cert.is_expiring_soon) {
              return false
            }
            break
          case "expiring":
            if (!cert.is_expiring_soon || cert.is_revoked || cert.is_expired) {
              return false
            }
            break
          case "expired":
            if (!cert.is_expired) {
              return false
            }
            break
          case "revoked":
            if (!cert.is_revoked) {
              return false
            }
            break
        }
      }

      return true
    })
  }, [certificates, filters])

  const handleDownload = async (serialNumber: string, commonName: string) => {
    try {
      setActionLoading(`download-${serialNumber}`)
      // Use the proxy endpoint to download certificate
      const response = await fetch(`/api/proxy/download/${encodeURIComponent(commonName)}/pem`, {
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
      a.download = `${commonName}.pem`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      toast({
        title: "Download successful",
        description: `Certificate ${commonName} downloaded successfully.`,
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

  const handleRenew = async (serialNumber: string, commonName: string) => {
    try {
      setActionLoading(`renew-${serialNumber}`)
      const result = await renewCertificate(serialNumber)
      
      if (result.success) {
        toast({
          title: "Certificate renewed",
          description: `Certificate ${commonName} was renewed successfully.`,
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

  const handleRevoke = async (serialNumber: string, commonName: string) => {
    if (!confirm(`Are you sure you want to revoke the certificate "${commonName}"? This action cannot be undone.`)) {
      return
    }
    
    try {
      setActionLoading(`revoke-${serialNumber}`)
      const result = await revokeCertificate(serialNumber)
      
      if (result.success) {
        toast({
          title: "Certificate revoked",
          description: `Certificate ${commonName} was revoked successfully.`,
        })
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

  const handleDelete = async (serialNumber: string, commonName: string) => {
    if (!confirm(`Are you sure you want to delete the certificate "${commonName}"? This action cannot be undone.`)) {
      return
    }
    
    try {
      setActionLoading(`delete-${serialNumber}`)
      const result = await deleteCertificate(serialNumber)
      
      if (result.success) {
        toast({
          title: "Certificate deleted",
          description: `Certificate ${commonName} was deleted successfully.`,
        })
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error.message}</AlertDescription>
      </Alert>
    )
  }

  if (filteredCertificates.length === 0) {
    return (
      <div className="rounded-md border p-8 text-center">
        <p className="text-muted-foreground">
          {certificates.length === 0 
            ? "No certificates found. Create your first certificate."
            : "No certificates match the current filters."
          }
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Common Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Expiry Date</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Serial Number</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filteredCertificates.map((certificate) => (
            <TableRow key={certificate.serial_number}>
              <TableCell className="font-medium">
                <div className="flex items-center space-x-2">
                  <FileText className="h-4 w-4 text-muted-foreground" />
                  <Link href={`/certificates/${certificate.serial_number}`} className="hover:underline">
                    {certificate.common_name}
                  </Link>
                </div>
              </TableCell>
              <TableCell>{certificate.is_client ? "Client" : "Server"}</TableCell>
              <TableCell>{certificate.expiry_date}</TableCell>
              <TableCell>
                {certificate.is_revoked ? (
                  <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                    <XCircle className="h-3 w-3" />
                    Revoked
                  </Badge>
                ) : certificate.is_expired ? (
                  <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                    <XCircle className="h-3 w-3" />
                    Expired
                  </Badge>
                ) : certificate.is_expiring_soon ? (
                  <Badge
                    variant="outline"
                    className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50"
                  >
                    <AlertTriangle className="h-3 w-3" />
                    Expires Soon
                  </Badge>
                ) : (
                  <Badge
                    variant="outline"
                    className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50"
                  >
                    <CheckCircle className="h-3 w-3" />
                    Valid
                  </Badge>
                )}
              </TableCell>
              <TableCell className="font-mono text-xs">{certificate.serial_number}</TableCell>
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem 
                      onClick={() => handleDownload(certificate.serial_number, certificate.common_name)}
                      disabled={actionLoading === `download-${certificate.serial_number}`}
                    >
                      {actionLoading === `download-${certificate.serial_number}` ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <Download className="mr-2 h-4 w-4" />
                      )}
                      <span>Download</span>
                    </DropdownMenuItem>
                    {!certificate.is_revoked && (
                      <DropdownMenuItem 
                        onClick={() => handleRenew(certificate.serial_number, certificate.common_name)}
                        disabled={actionLoading === `renew-${certificate.serial_number}`}
                      >
                        {actionLoading === `renew-${certificate.serial_number}` ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <RefreshCw className="mr-2 h-4 w-4" />
                        )}
                        <span>Renew</span>
                      </DropdownMenuItem>
                    )}
                    {!certificate.is_revoked && (
                      <DropdownMenuItem 
                        onClick={() => handleRevoke(certificate.serial_number, certificate.common_name)}
                        disabled={actionLoading === `revoke-${certificate.serial_number}`}
                      >
                        {actionLoading === `revoke-${certificate.serial_number}` ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <XCircle className="mr-2 h-4 w-4" />
                        )}
                        <span>Revoke</span>
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuSeparator />
                    <DropdownMenuItem 
                      onClick={() => handleDelete(certificate.serial_number, certificate.common_name)}
                      disabled={actionLoading === `delete-${certificate.serial_number}`}
                      className="text-red-600"
                    >
                      {actionLoading === `delete-${certificate.serial_number}` ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <XCircle className="mr-2 h-4 w-4" />
                      )}
                      <span>Delete</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
