import { DashboardHeader } from "@/components/dashboard-header"
import { CreateCertificateForm } from "@/components/create-certificate-form"
import { Breadcrumb } from "@/components/breadcrumb"

export default function CreateCertificatePage() {
  return (
    <div className="flex min-h-screen flex-col">
      <DashboardHeader />
      <main className="flex-1 space-y-4 p-8 pt-6">
        <Breadcrumb
          items={[
            { label: "Certificates", href: "/certificates" },
            { label: "Create Certificate", href: "/create" },
          ]}
        />
        <div className="flex flex-col space-y-6">
          <h2 className="text-3xl font-bold tracking-tight">Create Certificate</h2>
          <CreateCertificateForm />
        </div>
      </main>
    </div>
  )
}
