"use client"

import Link from "next/link"
import { useState } from "react"
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
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { Shield, Menu, Home, FileText, Plus, Settings, LogOut, Bell, Download, RefreshCw } from "lucide-react"
import { useToast } from "@/hooks/use-toast-new"
import { useCertificates } from "@/hooks/use-certificates"

export function DashboardHeader() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const router = useRouter()
  const { toast } = useToast()
  const { fetchCertificates, certificates } = useCertificates()

  const handleLogout = async () => {
    try {
      const response = await fetch('/api/proxy/logout', {
        method: 'GET',
        credentials: 'include',
      })
      
      if (response.ok) {
        toast({
          title: "Logged out successfully",
          description: "You have been logged out of LocalCA.",
        })
        router.push('/login')
      } else {
        throw new Error('Logout failed')
      }
    } catch (error) {
      toast({
        title: "Logout failed",
        description: "There was an error logging out. Please try again.",
        variant: "destructive",
      })
    }
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

  const handleDownloadCRL = async () => {
    try {
      const response = await fetch('/api/proxy/download/crl', {
        credentials: 'include'
      })
      
      if (!response.ok) {
        throw new Error('Failed to download CRL')
      }
      
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'ca.crl'
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      toast({
        title: "Download successful",
        description: "CRL downloaded successfully.",
      })
    } catch (error) {
      toast({
        title: "Download failed",
        description: error instanceof Error ? error.message : "Failed to download CRL",
        variant: "destructive",
      })
    }
  }

  const handleRefreshCertificates = async () => {
    try {
      await fetchCertificates()
      toast({
        title: "Refresh successful",
        description: "Certificate list refreshed successfully.",
      })
    } catch (error) {
      toast({
        title: "Refresh failed",
        description: error instanceof Error ? error.message : "Failed to refresh certificates",
        variant: "destructive",
      })
    }
  }

  // Get expiring certificates for notifications
  const expiringCertificates = certificates.filter(cert => cert.is_expiring_soon && !cert.is_revoked)
  const recentlyCreatedCertificates = certificates.slice(0, 3) // Show last 3 certificates as "recent"

  return (
    <header className="sticky top-0 z-40 border-b bg-background">
      <div className="container flex h-16 items-center justify-between py-4">
        <div className="flex items-center gap-2">
          <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" size="icon" className="md:hidden">
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-72">
              <nav className="grid gap-6 text-lg font-medium">
                <Link
                  href="/"
                  className="flex items-center gap-2 text-lg font-semibold"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <Shield className="h-6 w-6" />
                  <span>LocalCA</span>
                </Link>
                <Link
                  href="/"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <Home className="h-5 w-5" />
                  <span>Dashboard</span>
                </Link>
                <Link
                  href="/certificates"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <FileText className="h-5 w-5" />
                  <span>Certificates</span>
                </Link>
                <Link
                  href="/create"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <Plus className="h-5 w-5" />
                  <span>Create Certificate</span>
                </Link>
                <Link
                  href="/settings"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <Settings className="h-5 w-5" />
                  <span>Settings</span>
                </Link>
              </nav>
            </SheetContent>
          </Sheet>
          <Link href="/" className="flex items-center gap-2 md:hidden">
            <Shield className="h-6 w-6" />
            <span className="font-bold">LocalCA</span>
          </Link>
          <Link href="/" className="hidden items-center gap-2 md:flex">
            <Shield className="h-6 w-6" />
            <span className="text-xl font-bold">LocalCA</span>
          </Link>
          <nav className="hidden md:flex md:gap-6 md:text-sm md:font-medium md:ml-6">
            <Link href="/" className="transition-colors hover:text-foreground/80 text-sm font-medium">
              Dashboard
            </Link>
            <Link href="/certificates" className="transition-colors hover:text-foreground/80 text-sm font-medium">
              Certificates
            </Link>
            <Link href="/create" className="transition-colors hover:text-foreground/80 text-sm font-medium">
              Create Certificate
            </Link>
            <Link href="/settings" className="transition-colors hover:text-foreground/80 text-sm font-medium">
              Settings
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon" className="relative">
                <Bell className="h-5 w-5" />
                {expiringCertificates.length > 0 && (
                  <span className="absolute -top-1 -right-1 h-3 w-3 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
                    {expiringCertificates.length}
                  </span>
                )}
                <span className="sr-only">Notifications</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80">
              <DropdownMenuLabel>Notifications</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {expiringCertificates.length > 0 ? (
                expiringCertificates.map((cert) => (
                  <DropdownMenuItem key={cert.serial_number} asChild>
                    <Link href={`/certificates/${cert.serial_number}`}>
                      Certificate "{cert.common_name}" expires soon
                    </Link>
                  </DropdownMenuItem>
                ))
              ) : (
                <DropdownMenuItem disabled>
                  No notifications
                </DropdownMenuItem>
              )}
              {recentlyCreatedCertificates.length > 0 && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuLabel>Recent Activity</DropdownMenuLabel>
                  {recentlyCreatedCertificates.map((cert) => (
                    <DropdownMenuItem key={`recent-${cert.serial_number}`} asChild>
                      <Link href={`/certificates/${cert.serial_number}`}>
                        Certificate "{cert.common_name}" was created
                      </Link>
                    </DropdownMenuItem>
                  ))}
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon">
                <Download className="h-5 w-5" />
                <span className="sr-only">Download</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Download</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleDownloadCA}>
                CA Certificate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDownloadCRL}>
                CRL
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon">
                <RefreshCw className="h-5 w-5" />
                <span className="sr-only">Refresh</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Refresh</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleRefreshCertificates}>
                Refresh Certificate List
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDownloadCRL}>
                Refresh CRL
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" className="hidden md:flex">
                Admin
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>My Account</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/settings" className="flex items-center">
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Settings</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleLogout} className="cursor-pointer">
                <LogOut className="mr-2 h-4 w-4" />
                <span>Log out</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
