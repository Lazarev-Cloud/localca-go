import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateDetails } from "@/components/certificate-details"
import { CertificateActions } from "@/components/certificate-actions"
import { Breadcrumb } from "@/components/breadcrumb"

export default function CertificateDetailsPage({ params }: { params: { id: string } }) {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Breadcrumb
          items={[
            { label: "Certificates", href: "/certificates" },
            { label: "Certificate Details", href: `/certificates/${params.id}` },
          ]}
        />
        <div className="flex flex-col space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-bold tracking-tight">Certificate Details</h2>
            <CertificateActions id={params.id} />
          </div>
          <CertificateDetails id={params.id} />
        </div>
      </main>
    </div>
  )
}
