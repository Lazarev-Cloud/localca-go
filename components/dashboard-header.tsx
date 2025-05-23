"use client"

import Link from "next/link"
import { useState, useCallback } from "react"
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
import { Shield, Menu, Home, FileText, Plus, Settings, LogOut, Bell, Download, RefreshCw, AlertTriangle, Clock, CheckCircle } from "lucide-react"
import { useToast } from "@/hooks/use-toast-new"
import { useCertificates } from "@/hooks/use-certificates"
import { useAuditLogs } from "@/hooks/use-audit-logs"

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

export function DashboardHeader() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const router = useRouter()
  const { toast } = useToast()
  const { fetchCertificates, certificates } = useCertificates()
  const { auditLogs: recentActivity } = useAuditLogs(5, 0)

  const handleLogout = async () => {
    try {
      const response = await fetch('/api/logout', {
        method: 'POST',
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
      const response = await fetch('/api/proxy/api/download/ca', {
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
      const response = await fetch('/api/proxy/api/download/crl', {
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
  
  // Calculate days until expiry
  const getDaysUntilExpiry = (expiryDate: string) => {
    const expiry = new Date(expiryDate)
    const now = new Date()
    const diffInMs = expiry.getTime() - now.getTime()
    const diffInDays = Math.ceil(diffInMs / (1000 * 60 * 60 * 24))
    return diffInDays
  }

  // Format time ago for activity
  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)
    
    if (diffInSeconds < 60) {
      return 'Just now'
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60)
      return `${minutes}m ago`
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600)
      return `${hours}h ago`
    } else {
      const days = Math.floor(diffInSeconds / 86400)
      return `${days}d ago`
    }
  }

  // Get activity icon
  const getActivityIcon = (action: string, success: boolean) => {
    if (!success) {
      return <AlertTriangle className="h-4 w-4 text-red-500" />
    }
    
    switch (action.toLowerCase()) {
      case 'create':
        return <Plus className="h-4 w-4 text-green-500" />
      case 'download':
        return <Download className="h-4 w-4 text-blue-500" />
      case 'revoke':
      case 'delete':
        return <AlertTriangle className="h-4 w-4 text-red-500" />
      default:
        return <CheckCircle className="h-4 w-4 text-green-500" />
    }
  }

  // Total notifications count
  const totalNotifications = expiringCertificates.length + recentActivity.length

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
                {totalNotifications > 0 && (
                  <span className="absolute -top-1 -right-1 h-4 w-4 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
                    {totalNotifications > 9 ? '9+' : totalNotifications}
                  </span>
                )}
                <span className="sr-only">Notifications</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-96 max-h-96 overflow-y-auto">
              <DropdownMenuLabel className="flex items-center gap-2">
                <Bell className="h-4 w-4" />
                Notifications
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              
              {/* Expiring Certificates */}
              {expiringCertificates.length > 0 && (
                <>
                  <div className="px-2 py-1">
                    <div className="flex items-center gap-2 text-sm font-medium text-orange-600">
                      <Clock className="h-4 w-4" />
                      Expiring Soon
                    </div>
                  </div>
                  {expiringCertificates.slice(0, 3).map((cert) => (
                    <DropdownMenuItem key={cert.serial_number} asChild className="cursor-pointer">
                      <Link href="/certificates" className="flex items-start gap-3 p-3">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5 flex-shrink-0" />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">
                            Certificate expiring soon
                          </p>
                          <p className="text-xs text-muted-foreground truncate">
                            "{cert.common_name}" expires {getDaysUntilExpiry(cert.expiry_date)} days
                          </p>
                        </div>
                      </Link>
                    </DropdownMenuItem>
                  ))}
                  {expiringCertificates.length > 3 && (
                    <DropdownMenuItem asChild className="cursor-pointer">
                      <Link href="/certificates" className="text-center text-sm text-muted-foreground">
                        View {expiringCertificates.length - 3} more expiring certificates
                      </Link>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                </>
              )}

              {/* Recent Activity */}
              {recentActivity.length > 0 && (
                <>
                  <div className="px-2 py-1">
                    <div className="flex items-center gap-2 text-sm font-medium text-blue-600">
                      <RefreshCw className="h-4 w-4" />
                      Recent Activity
                    </div>
                  </div>
                  {recentActivity.slice(0, 3).map((activity) => (
                    <DropdownMenuItem key={activity.id} className="cursor-pointer">
                      <div className="flex items-start gap-3 p-3 w-full">
                        {getActivityIcon(activity.action, activity.success)}
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">
                            {activity.action.charAt(0).toUpperCase() + activity.action.slice(1)} {activity.resource}
                            {!activity.success && <span className="text-red-500 ml-1">(Failed)</span>}
                          </p>
                          <p className="text-xs text-muted-foreground truncate">
                            {activity.resource_id && `"${activity.resource_id}"`} • {formatTimeAgo(activity.created_at)}
                          </p>
                        </div>
                      </div>
                    </DropdownMenuItem>
                  ))}
                  {recentActivity.length > 3 && (
                    <DropdownMenuItem asChild className="cursor-pointer">
                      <Link href="/" className="text-center text-sm text-muted-foreground">
                        View all activity
                      </Link>
                    </DropdownMenuItem>
                  )}
                </>
              )}

              {/* No notifications */}
              {totalNotifications === 0 && (
                <div className="p-6 text-center">
                  <Bell className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                  <p className="text-sm text-muted-foreground">No notifications</p>
                  <p className="text-xs text-muted-foreground mt-1">You're all caught up!</p>
                </div>
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
