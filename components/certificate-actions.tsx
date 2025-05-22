"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Download, RefreshCw, XCircle, MoreHorizontal } from "lucide-react"
import { revokeCertificate, renewCertificate, getCertificateDownloadUrl } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"
import { useRouter } from "next/navigation"

interface CertificateActionsProps {
  id: string
}

export function CertificateActions({ id }: CertificateActionsProps) {
  const [isLoading, setIsLoading] = useState(false)
  const { toast } = useToast()
  const router = useRouter()

  // Handle certificate revocation
  const handleRevoke = async () => {
    try {
      setIsLoading(true)
      const response = await revokeCertificate(id)

      if (response.success) {
        toast({
          title: "Certificate Revoked",
          description: "The certificate has been successfully revoked.",
        })

        // Refresh the page to show updated status
        router.refresh()
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
      setIsLoading(false)
    }
  }

  // Handle certificate renewal
  const handleRenew = async () => {
    try {
      setIsLoading(true)
      const response = await renewCertificate(id)

      if (response.success) {
        toast({
          title: "Certificate Renewed",
          description: "The certificate has been successfully renewed.",
        })

        // Refresh the page to show updated expiry date
        router.refresh()
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
      setIsLoading(false)
    }
  }

  // Handle certificate download
  const handleDownload = (format: "pem" | "p12" | "key" = "pem") => {
    const downloadUrl = getCertificateDownloadUrl(id, format)
    window.open(downloadUrl, "_blank")
  }

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        className="flex items-center gap-2"
        onClick={() => handleDownload("pem")}
        disabled={isLoading}
      >
        <Download className="h-4 w-4" />
        Download
      </Button>
      <Button variant="outline" className="flex items-center gap-2" onClick={handleRenew} disabled={isLoading}>
        <RefreshCw className="h-4 w-4" />
        Renew
      </Button>
      <Button
        variant="outline"
        className="flex items-center gap-2 text-red-500 hover:text-red-600"
        onClick={handleRevoke}
        disabled={isLoading}
      >
        <XCircle className="h-4 w-4" />
        Revoke
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon" disabled={isLoading}>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>More Actions</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => handleDownload("pem")}>Download as PEM</DropdownMenuItem>
          <DropdownMenuItem onClick={() => handleDownload("p12")}>Download as P12</DropdownMenuItem>
          <DropdownMenuItem onClick={() => handleDownload("key")}>Download Private Key</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => router.push("/certificates")}>View All Certificates</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
