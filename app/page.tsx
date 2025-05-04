import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateList } from "@/components/certificate-list"
import { CAInfoCard } from "@/components/ca-info-card"
import { QuickActions } from "@/components/quick-actions"
import { RecentActivity } from "@/components/recent-activity"
import { SystemStatus } from "@/components/system-status"

export default async function DashboardPage() {
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
    </div>
  )
}
