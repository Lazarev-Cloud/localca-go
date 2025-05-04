import { DashboardHeader } from "@/components/dashboard-header"
import { SettingsTabs } from "@/components/settings-tabs"
import { Breadcrumb } from "@/components/breadcrumb"

export default function SettingsPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Breadcrumb items={[{ label: "Settings", href: "/settings" }]} />
        <div className="flex flex-col space-y-6">
          <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
          <SettingsTabs />
        </div>
      </main>
    </div>
  )
}
