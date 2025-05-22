import { Suspense } from "react"
import { DashboardHeader } from "@/components/dashboard-header"
import { SettingsTabs } from "@/components/settings-tabs"
import { Breadcrumb } from "@/components/breadcrumb"
import { Loading } from "@/components/ui/loading"
import { ErrorDisplay } from "@/components/ui/error"
import { getSettings } from "@/lib/api"

async function SettingsContent() {
  try {
    // Fetch settings
    const settings = await getSettings()

    return (
      <>
        <Breadcrumb items={[{ label: "Settings", href: "/settings" }]} />
        <div className="flex flex-col space-y-6">
          <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
          <SettingsTabs initialSettings={settings} />
        </div>
      </>
    )
  } catch (error) {
    return (
      <ErrorDisplay
        title="Failed to load settings"
        message={error instanceof Error ? error.message : "An unknown error occurred"}
      />
    )
  }
}

export default function SettingsPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Suspense fallback={<Loading message="Loading settings..." />}>
          <SettingsContent />
        </Suspense>
      </main>
    </div>
  )
}
