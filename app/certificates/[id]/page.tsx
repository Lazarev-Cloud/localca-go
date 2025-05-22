import { Suspense } from "react"
import { DashboardHeader } from "@/components/dashboard-header"
import { CertificateDetails } from "@/components/certificate-details"
import { CertificateActions } from "@/components/certificate-actions"
import { Breadcrumb } from "@/components/breadcrumb"
import { Loading } from "@/components/ui/loading"
import { ErrorDisplay } from "@/components/ui/error"
import { getCertificate } from "@/lib/api"
import { notFound } from "next/navigation"

async function CertificateDetailsContent({ id }: { id: string }) {
  try {
    // Fetch certificate details
    const certificate = await getCertificate(id)

    return (
      <>
        <Breadcrumb
          items={[
            { label: "Certificates", href: "/certificates" },
            { label: "Certificate Details", href: `/certificates/${id}` },
          ]}
        />
        <div className="flex flex-col space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-bold tracking-tight">Certificate Details</h2>
            <CertificateActions id={id} />
          </div>
          <CertificateDetails certificate={certificate} />
        </div>
      </>
    )
  } catch (error) {
    if (error instanceof Error && "status" in error && error.status === 404) {
      notFound()
    }

    return (
      <ErrorDisplay
        title="Failed to load certificate details"
        message={error instanceof Error ? error.message : "An unknown error occurred"}
      />
    )
  }
}

export default function CertificateDetailsPage({ params }: { params: { id: string } }) {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Suspense fallback={<Loading message="Loading certificate details..." />}>
          <CertificateDetailsContent id={params.id} />
        </Suspense>
      </main>
    </div>
  )
}
