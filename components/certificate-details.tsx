"use client"

import { useState, useEffect } from "react"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AlertTriangle, CheckCircle, XCircle, Calendar, FileText, Key, Globe, Building, Hash, Loader2 } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { useCertificates } from "@/hooks/use-certificates"

interface CertificateDetailsProps {
  id: string
}

export function CertificateDetails({ id }: CertificateDetailsProps) {
  const { certificates, loading, error } = useCertificates()
  const [certificate, setCertificate] = useState<any>(null)

  useEffect(() => {
    // Find the certificate by serial number
    const cert = certificates.find(c => c.serial_number === id)
    setCertificate(cert)
  }, [certificates, id])

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error.message}</AlertDescription>
      </Alert>
    )
  }

  if (!certificate) {
    return (
      <Alert>
        <AlertDescription>Certificate not found.</AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <CardTitle className="text-2xl">{certificate.common_name}</CardTitle>
            <CardDescription>{certificate.is_client ? "Client" : "Server"} Certificate</CardDescription>
          </div>
          <div>
            {certificate.is_revoked ? (
              <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                <XCircle className="h-3 w-3" />
                Revoked
              </Badge>
            ) : certificate.is_expired ? (
              <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                <XCircle className="h-3 w-3" />
                Expired
              </Badge>
            ) : certificate.is_expiring_soon ? (
              <Badge variant="outline" className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50">
                <AlertTriangle className="h-3 w-3" />
                Expires Soon
              </Badge>
            ) : (
              <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                <CheckCircle className="h-3 w-3" />
                Valid
              </Badge>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="general">
            <TabsList>
              <TabsTrigger value="general">General</TabsTrigger>
              <TabsTrigger value="details">Details</TabsTrigger>
            </TabsList>
            <TabsContent value="general" className="space-y-4">
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <FileText className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Common Name</p>
                    <p className="text-sm text-muted-foreground">{certificate.common_name}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Certificate Type</p>
                    <p className="text-sm text-muted-foreground">{certificate.is_client ? "Client Certificate" : "Server Certificate"}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Expiry Date</p>
                    <p className="text-sm text-muted-foreground">{certificate.expiry_date}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Hash className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Serial Number</p>
                    <p className="text-sm font-mono text-muted-foreground">{certificate.serial_number}</p>
                  </div>
                </div>
              </div>
            </TabsContent>
            <TabsContent value="details" className="space-y-4">
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <CheckCircle className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Status</p>
                    <p className="text-sm text-muted-foreground">
                      {certificate.is_revoked ? "Revoked" : 
                       certificate.is_expired ? "Expired" : 
                       certificate.is_expiring_soon ? "Expiring Soon" : "Valid"}
                    </p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <FileText className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Certificate Usage</p>
                    <p className="text-sm text-muted-foreground">
                      {certificate.is_client ? "Client Authentication" : "Server Authentication"}
                    </p>
                  </div>
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}

