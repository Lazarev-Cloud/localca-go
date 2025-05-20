"use client"

import Link from "next/link"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { FileText, AlertTriangle, CheckCircle, XCircle, Loader2 } from "lucide-react"
import { useCertificates } from "@/hooks/use-certificates"
import { Alert, AlertDescription } from "@/components/ui/alert"

interface CertificateListProps {
  className?: string
}

export function CertificateList({ className }: CertificateListProps) {
  const { certificates, loading, error } = useCertificates()

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
        {loading && (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {!loading && !error && certificates.length === 0 && (
          <div className="py-8 text-center text-muted-foreground">
            No certificates found. Create your first certificate.
          </div>
        )}

        {!loading && !error && certificates.length > 0 && (
          <div className="space-y-4">
            {certificates.slice(0, 5).map((cert) => (
              <div key={cert.serial_number} className="flex items-center justify-between space-x-4">
                <div className="flex items-center space-x-4">
                  <FileText className="h-6 w-6 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium leading-none">{cert.common_name}</p>
                    <p className="text-sm text-muted-foreground">
                      {cert.is_client ? "Client Certificate" : "Server Certificate"}
                    </p>
                  </div>
                </div>
                <div className="flex items-center">
                  {cert.is_revoked && (
                    <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                      <XCircle className="h-3 w-3" />
                      Revoked
                    </Badge>
                  )}
                  {!cert.is_revoked && cert.is_expired && (
                    <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                      <XCircle className="h-3 w-3" />
                      Expired
                    </Badge>
                  )}
                  {!cert.is_revoked && !cert.is_expired && cert.is_expiring_soon && (
                    <Badge variant="outline" className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50">
                      <AlertTriangle className="h-3 w-3" />
                      Expires Soon
                    </Badge>
                  )}
                  {!cert.is_revoked && !cert.is_expired && !cert.is_expiring_soon && (
                    <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                      <CheckCircle className="h-3 w-3" />
                      Valid
                    </Badge>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
