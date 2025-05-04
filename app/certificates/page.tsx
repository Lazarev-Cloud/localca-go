import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateTable } from "@/components/certificate-table"
import { CertificateFilters } from "@/components/certificate-filters"
import { CreateCertificateButton } from "@/components/create-certificate-button"

export default function CertificatesPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <div className="flex items-center justify-between">
          <h2 className="text-3xl font-bold tracking-tight">Certificates</h2>
          <CreateCertificateButton />
        </div>
        <div className="space-y-4">
          <CertificateFilters />
          <CertificateTable />
        </div>
      </main>
    </div>
  )
}
