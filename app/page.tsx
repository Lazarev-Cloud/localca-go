import { Suspense } from "react"
import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateList } from "@/components/certificate-list"
import { CAInfoCard } from "@/components/ca-info-card"
import { QuickActions } from "@/components/quick-actions"
import { RecentActivity } from "@/components/recent-activity"
import { SystemStatus } from "@/components/system-status"
import { Loading } from "@/components/ui/loading"
import { ErrorDisplay } from "@/components/ui/error"
import { getCAInfo, getSystemStatus, getCertificates, getRecentActivity } from "@/lib/api"

// Dashboard content with error handling
async function DashboardContent() {
  try {
    // Fetch all data in parallel
    const [caInfo, systemStatus, certificates, activities] = await Promise.all([
      getCAInfo(),
      getSystemStatus(),
      getCertificates(),
      getRecentActivity(),
    ])

    return (
      <>
        <div className="flex items-center justify-between space-y-2">
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <div className="flex items-center space-x-2">
            <QuickActions />
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <CAInfoCard className="col-span-3" caInfo={caInfo} />
          <SystemStatus className="col-span-4" systemStatus={systemStatus} />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <CertificateList className="col-span-4" certificates={certificates.slice(0, 5)} />
          <RecentActivity className="col-span-3" activities={activities} />
        </div>
      </>
    )
  } catch (error) {
    return (
      <ErrorDisplay
        title="Failed to load dashboard data"
        message={error instanceof Error ? error.message : "An unknown error occurred"}
      />
    )
  }
}

export default function DashboardPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Suspense fallback={<Loading message="Loading dashboard data..." />}>
          <DashboardContent />
        </Suspense>
      </main>
    </div>
  )
}
