"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Shield, AlertTriangle, CheckCircle, Loader2 } from "lucide-react"
import { useCertificates } from "@/hooks/use-certificates"
import { Alert, AlertDescription } from "@/components/ui/alert"

interface CAInfoCardProps {
  className?: string
}

export function CAInfoCard({ className }: CAInfoCardProps) {
  const { caInfo, loading, error } = useCertificates()

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle>Certificate Authority</CardTitle>
          <CardDescription>Information about your CA</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        {loading && (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error.message}</AlertDescription>
          </Alert>
        )}

        {!loading && !error && !caInfo && (
          <div className="py-8 text-center text-muted-foreground">
            CA information not available.
          </div>
        )}

        {!loading && !error && caInfo && (
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <Shield className="h-10 w-10 text-primary" />
              <div>
                <p className="text-lg font-medium">{caInfo.common_name}</p>
                <div className="flex items-center space-x-2">
                  {caInfo.is_expired ? (
                    <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                      <AlertTriangle className="h-3 w-3" />
                      Expired
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                      <CheckCircle className="h-3 w-3" />
                      Valid
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium">Organization</p>
                <p className="text-sm text-muted-foreground">{caInfo.organization}</p>
              </div>
              <div>
                <p className="text-sm font-medium">Country</p>
                <p className="text-sm text-muted-foreground">{caInfo.country}</p>
              </div>
              <div>
                <p className="text-sm font-medium">Expiry Date</p>
                <p className="text-sm text-muted-foreground">{caInfo.expiry_date}</p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
