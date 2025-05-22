import Link from "next/link"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { FileText, AlertTriangle, CheckCircle, XCircle } from "lucide-react"
import type { Certificate } from "@/lib/api"

interface CertificateListProps {
  className?: string
  certificates: Certificate[]
}

export function CertificateList({ className, certificates }: CertificateListProps) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle>Recent Certificates</CardTitle>
          <CardDescription>Recently created or updated certificates</CardDescription>
        </div>
        <Link href="/certificates" className="text-sm text-muted-foreground hover:underline">
          View all
        </Link>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {certificates.length > 0 ? (
            certificates.map((cert) => (
              <div key={cert.id} className="flex items-center justify-between space-x-4">
                <div className="flex items-center space-x-4">
                  <FileText className="h-6 w-6 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium leading-none">{cert.common_name}</p>
                    <p className="text-sm text-muted-foreground">{cert.type} Certificate</p>
                  </div>
                </div>
                <div className="flex items-center">
                  {cert.is_revoked ? (
                    <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                      <XCircle className="h-3 w-3" />
                      Revoked
                    </Badge>
                  ) : cert.is_expired ? (
                    <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                      <XCircle className="h-3 w-3" />
                      Expired
                    </Badge>
                  ) : cert.is_expiring_soon ? (
                    <Badge
                      variant="outline"
                      className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50"
                    >
                      <AlertTriangle className="h-3 w-3" />
                      Expires Soon
                    </Badge>
                  ) : (
                    <Badge
                      variant="outline"
                      className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50"
                    >
                      <CheckCircle className="h-3 w-3" />
                      Valid
                    </Badge>
                  )}
                </div>
              </div>
            ))
          ) : (
            <div className="flex items-center justify-center p-4 text-muted-foreground">No certificates found</div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
