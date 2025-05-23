'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import config from '@/lib/config'
import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateList } from "@/components/certificate-list"
import { CAInfoCard } from "@/components/ca-info-card"
import { QuickActions } from "@/components/quick-actions"
import { RecentActivity } from "@/components/recent-activity"
import { SystemStatus } from "@/components/system-status"
import { SetupRedirect } from '@/components/setup-redirect'

export default function Home() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [debugInfo, setDebugInfo] = useState<any>(null)

  useEffect(() => {
    async function checkSetupStatus() {
      try {
        console.log('Checking authentication status...')
        
        // Try the direct proxy endpoint for best cookie forwarding
        console.log('Trying proxy endpoint for CA info...')
        const proxyResponse = await fetch('/api/proxy/ca-info', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Cache-Control': 'no-cache'
          }
        })
        
        console.log('Proxy endpoint response status:', proxyResponse.status)
        
        // If successful with proxy, show the dashboard
        if (proxyResponse.ok) {
          console.log('Proxy endpoint successful, showing dashboard')
          setLoading(false)
          return
        }
        
        // Process the response for redirection if needed
        if (proxyResponse.status === 401) {
          try {
            const data = await proxyResponse.json()
            console.log('Proxy 401 response data:', data)
            
            if (data.setupRequired || 
                (data.data && data.data.setup_required) ||
                data.message === 'Setup required') {
              console.log('Setup required, redirecting to setup page')
              router.push('/setup')
            } else {
              console.log('Authentication required, redirecting to login')
              router.push('/login')
            }
          } catch (err) {
            console.error('Error parsing proxy response:', err)
            // Default to login
            router.push('/login')
          }
          return
        }
        
        // Fallback to debug info for troubleshooting
        const debugResponse = await fetch('/api/debug-ca-info', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Cache-Control': 'no-cache'
          }
        })
        
        if (debugResponse.ok) {
          const debugData = await debugResponse.json()
          console.log('Debug API response:', debugData)
          setDebugInfo(debugData)
        }
        
        // Check the setup status by calling the CA info endpoint
        const response = await fetch('/api/ca-info', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Cache-Control': 'no-cache'
          }
        })

        console.log('CA info response status:', response.status)
        
        if (response.status === 401) {
          // If 401, check if setup is required
          const data = await response.json()
          console.log('401 response data:', data)
          
          if (data.setupRequired) {
            console.log('Setup required, redirecting to setup page')
            // Setup required, redirect to setup page
            router.push('/setup')
          } else {
            console.log('Authentication required, redirecting to login')
            // Authentication required, redirect to login
            router.push('/login')
          }
        } else if (response.ok) {
          console.log('Already authenticated, staying on home page')
          // Already authenticated and set up, stay on home page
          setLoading(false)
        } else {
          console.log('Other error, redirecting to login')
          // Other error, try to login
          router.push('/login')
        }
      } catch (error) {
        console.error('Error checking setup status:', error)
        // On error, redirect to login
        router.push('/login')
      }
    }

    checkSetupStatus()
  }, [router])

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold">Loading LocalCA...</h1>
          <p className="mt-2 text-gray-500">Please wait while we check your setup status</p>
          {debugInfo && (
            <div className="mt-4 text-xs text-left overflow-auto max-h-[400px] max-w-[600px] bg-gray-100 p-2 rounded">
              <pre>{JSON.stringify(debugInfo, null, 2)}</pre>
            </div>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <div className="flex items-center justify-between space-y-2">
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <div className="flex items-center space-x-2">
            <QuickActions />
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <CAInfoCard className="col-span-3" />
          <SystemStatus className="col-span-4" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <CertificateList className="col-span-4" />
          <RecentActivity className="col-span-3" />
        </div>
      </main>
      <SetupRedirect />
    </div>
  )
}
