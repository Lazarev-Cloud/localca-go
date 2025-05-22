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

  useEffect(() => {
    async function checkSetupStatus() {
      try {
        // Check the setup status by calling the CA info endpoint
        const response = await fetch('/api/ca-info', {
          method: 'GET',
          credentials: 'include',
        })

        if (response.status === 401) {
          // If 401, check if setup is required
          const data = await response.json()
          if (data.setupRequired) {
            // Setup required, redirect to setup page
            router.push('/setup')
          } else {
            // Authentication required, redirect to login
            router.push('/login')
          }
        } else if (response.ok) {
          // Already authenticated and set up, stay on home page
          setLoading(false)
        } else {
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
