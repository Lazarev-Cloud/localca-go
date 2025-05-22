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
import { useToast } from "@/hooks/use-toast"

export function DashboardHeader() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const router = useRouter()
  const { toast } = useToast()

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
              <Button variant="outline" size="icon">
                <Bell className="h-5 w-5" />
                <span className="sr-only">Notifications</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Notifications</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Certificate "server.local" expires in 7 days</DropdownMenuItem>
              <DropdownMenuItem>Certificate "client.p12" was created successfully</DropdownMenuItem>
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
              <DropdownMenuItem>CA Certificate</DropdownMenuItem>
              <DropdownMenuItem>CA Certificate Chain</DropdownMenuItem>
              <DropdownMenuItem>CRL</DropdownMenuItem>
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
              <DropdownMenuItem>Refresh Certificate List</DropdownMenuItem>
              <DropdownMenuItem>Refresh CRL</DropdownMenuItem>
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
