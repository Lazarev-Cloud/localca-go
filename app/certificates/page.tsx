import { Suspense } from "react"
import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateTable } from "@/components/certificate-table"
import { CertificateFilters } from "@/components/certificate-filters"
import { CreateCertificateButton } from "@/components/create-certificate-button"
import { Loading } from "@/components/ui/loading"
import { ErrorDisplay } from "@/components/ui/error"
import { getCertificates } from "@/lib/api"

async function CertificatesContent() {
  try {
    // Fetch certificates
    const certificates = await getCertificates()

    return (
      <>
        <div className="flex items-center justify-between">
          <h2 className="text-3xl font-bold tracking-tight">Certificates</h2>
          <CreateCertificateButton />
        </div>
        <div className="space-y-4">
          <CertificateFilters />
          <CertificateTable initialCertificates={certificates} />
        </div>
      </>
    )
  } catch (error) {
    return (
      <ErrorDisplay
        title="Failed to load certificates"
        message={error instanceof Error ? error.message : "An unknown error occurred"}
      />
    )
  }
}

export default function CertificatesPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Suspense fallback={<Loading message="Loading certificates..." />}>
          <CertificatesContent />
        </Suspense>
      </main>
    </div>
  )
}
