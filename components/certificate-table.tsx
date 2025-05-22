"use client"

import { useState } from "react"
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
import { MoreHorizontal, Download, RefreshCw, XCircle, AlertTriangle, CheckCircle, FileText } from "lucide-react"
import { type Certificate, revokeCertificate, renewCertificate, getCertificateDownloadUrl } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"

interface CertificateTableProps {
  initialCertificates: Certificate[]
}

export function CertificateTable({ initialCertificates }: CertificateTableProps) {
  const [certificates, setCertificates] = useState<Certificate[]>(initialCertificates)
  const [isLoading, setIsLoading] = useState<Record<string, boolean>>({})
  const { toast } = useToast()

  // Handle certificate revocation
  const handleRevoke = async (id: string) => {
    try {
      setIsLoading((prev) => ({ ...prev, [id]: true }))
      const response = await revokeCertificate(id)

      if (response.success) {
        toast({
          title: "Certificate Revoked",
          description: "The certificate has been successfully revoked.",
        })

        // Update the certificate status in the UI
        setCertificates((prev) => prev.map((cert) => (cert.id === id ? { ...cert, is_revoked: true } : cert)))
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to revoke certificate.",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "An unknown error occurred",
        variant: "destructive",
      })
    } finally {
      setIsLoading((prev) => ({ ...prev, [id]: false }))
    }
  }

  // Handle certificate renewal
  const handleRenew = async (id: string) => {
    try {
      setIsLoading((prev) => ({ ...prev, [id]: true }))
      const response = await renewCertificate(id)

      if (response.success) {
        toast({
          title: "Certificate Renewed",
          description: "The certificate has been successfully renewed.",
        })

        // Update the certificate in the UI
        setCertificates((prev) =>
          prev.map((cert) => (cert.id === id ? { ...cert, is_expired: false, is_expiring_soon: false } : cert)),
        )
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to renew certificate.",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "An unknown error occurred",
        variant: "destructive",
      })
    } finally {
      setIsLoading((prev) => ({ ...prev, [id]: false }))
    }
  }

  // Handle certificate download
  const handleDownload = (id: string, format: "pem" | "p12" | "key" = "pem") => {
    const downloadUrl = getCertificateDownloadUrl(id, format)
    window.open(downloadUrl, "_blank")
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
          {certificates.length > 0 ? (
            certificates.map((certificate) => (
              <TableRow key={certificate.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center space-x-2">
                    <FileText className="h-4 w-4 text-muted-foreground" />
                    <Link href={`/certificates/${certificate.id}`} className="hover:underline">
                      {certificate.common_name}
                    </Link>
                  </div>
                </TableCell>
                <TableCell>{certificate.type}</TableCell>
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
                      <Button variant="ghost" className="h-8 w-8 p-0" disabled={isLoading[certificate.id]}>
                        <span className="sr-only">Open menu</span>
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => handleDownload(certificate.id, "pem")}>
                        <Download className="mr-2 h-4 w-4" />
                        <span>Download PEM</span>
                      </DropdownMenuItem>
                      {certificate.type === "Client" && (
                        <DropdownMenuItem onClick={() => handleDownload(certificate.id, "p12")}>
                          <Download className="mr-2 h-4 w-4" />
                          <span>Download P12</span>
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem onClick={() => handleDownload(certificate.id, "key")}>
                        <Download className="mr-2 h-4 w-4" />
                        <span>Download Key</span>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      {!certificate.is_revoked && !certificate.is_expired && (
                        <DropdownMenuItem
                          onClick={() => handleRenew(certificate.id)}
                          disabled={isLoading[certificate.id]}
                        >
                          <RefreshCw className="mr-2 h-4 w-4" />
                          <span>Renew</span>
                        </DropdownMenuItem>
                      )}
                      {!certificate.is_revoked && (
                        <DropdownMenuItem
                          onClick={() => handleRevoke(certificate.id)}
                          disabled={isLoading[certificate.id]}
                          className="text-red-500 focus:text-red-500"
                        >
                          <XCircle className="mr-2 h-4 w-4" />
                          <span>Revoke</span>
                        </DropdownMenuItem>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={6} className="h-24 text-center">
                No certificates found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
