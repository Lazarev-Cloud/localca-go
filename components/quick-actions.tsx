"use client"

import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Plus, Download, RefreshCw, Settings } from "lucide-react"
import { useToast } from "@/hooks/use-toast-new"
import { useCertificates } from "@/hooks/use-certificates"

export function QuickActions() {
  const router = useRouter()
  const { toast } = useToast()
  const { fetchCertificates } = useCertificates()

  const handleCreateCertificate = () => {
    router.push('/create')
  }

  const handleDownloadCA = async () => {
    try {
      const response = await fetch('/api/proxy/download/ca', {
        credentials: 'include'
      })
      
      if (!response.ok) {
        throw new Error('Failed to download CA certificate')
      }
      
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'ca.pem'
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      toast({
        title: "Download successful",
        description: "CA certificate downloaded successfully.",
      })
    } catch (error) {
      toast({
        title: "Download failed",
        description: error instanceof Error ? error.message : "Failed to download CA certificate",
        variant: "destructive",
      })
    }
  }

  const handleRefreshCRL = async () => {
    try {
      // Refresh the certificate list which will also update the CRL
      await fetchCertificates()
      
      toast({
        title: "Refresh successful",
        description: "Certificate list and CRL refreshed successfully.",
      })
    } catch (error) {
      toast({
        title: "Refresh failed",
        description: error instanceof Error ? error.message : "Failed to refresh CRL",
        variant: "destructive",
      })
    }
  }

  const handleSettings = () => {
    router.push('/settings')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Quick Actions
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>Actions</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleCreateCertificate}>
          <Plus className="mr-2 h-4 w-4" />
          <span>Create Certificate</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleDownloadCA}>
          <Download className="mr-2 h-4 w-4" />
          <span>Download CA Certificate</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleRefreshCRL}>
          <RefreshCw className="mr-2 h-4 w-4" />
          <span>Refresh CRL</span>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleSettings}>
          <Settings className="mr-2 h-4 w-4" />
          <span>Settings</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
