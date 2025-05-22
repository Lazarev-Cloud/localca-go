import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AlertTriangle, CheckCircle, XCircle, Calendar, FileText, Key, Globe, Building, Hash } from "lucide-react"
import type { Certificate } from "@/lib/api"

interface CertificateDetailsProps {
  certificate: Certificate
}

export function CertificateDetails({ certificate }: CertificateDetailsProps) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <CardTitle className="text-2xl">{certificate.common_name}</CardTitle>
            <CardDescription>{certificate.type} Certificate</CardDescription>
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
              <TabsTrigger value="extensions">Extensions</TabsTrigger>
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
                    <p className="text-sm font-medium leading-none">Valid From</p>
                    <p className="text-sm text-muted-foreground">{certificate.issued_date}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Valid Until</p>
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
                  <Building className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Organization</p>
                    <p className="text-sm text-muted-foreground">{certificate.organization || "Not specified"}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Globe className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Country</p>
                    <p className="text-sm text-muted-foreground">{certificate.country || "Not specified"}</p>
                  </div>
                </div>
                {certificate.type === "Server" &&
                  certificate.alternative_names &&
                  certificate.alternative_names.length > 0 && (
                    <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                      <Globe className="h-5 w-5 text-muted-foreground" />
                      <div className="space-y-1">
                        <p className="text-sm font-medium leading-none">Alternative Names</p>
                        <div className="flex flex-wrap gap-2 mt-1">
                          {certificate.alternative_names.map((name, index) => (
                            <Badge key={index} variant="secondary">
                              {name}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    </div>
                  )}
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Key className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Key Information</p>
                    <p className="text-sm text-muted-foreground">
                      {certificate.key_type || "RSA"}, {certificate.key_size || "2048 bits"}
                    </p>
                  </div>
                </div>
              </div>
            </TabsContent>
            <TabsContent value="extensions" className="space-y-4">
              <div className="grid gap-4 py-4">
                {certificate.key_usage && certificate.key_usage.length > 0 && (
                  <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                    <Key className="h-5 w-5 text-muted-foreground" />
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none">Key Usage</p>
                      <div className="flex flex-wrap gap-2 mt-1">
                        {certificate.key_usage.map((usage, index) => (
                          <Badge key={index} variant="secondary">
                            {usage}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
                {certificate.extended_key_usage && certificate.extended_key_usage.length > 0 && (
                  <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                    <Key className="h-5 w-5 text-muted-foreground" />
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none">Extended Key Usage</p>
                      <div className="flex flex-wrap gap-2 mt-1">
                        {certificate.extended_key_usage.map((usage, index) => (
                          <Badge key={index} variant="secondary">
                            {usage}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Hash className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Signature Algorithm</p>
                    <p className="text-sm text-muted-foreground">
                      {certificate.signature_algorithm || "SHA256withRSA"}
                    </p>
                  </div>
                </div>
                {certificate.fingerprint && (
                  <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                    <Hash className="h-5 w-5 text-muted-foreground" />
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none">Fingerprint (SHA-1)</p>
                      <p className="text-sm font-mono text-muted-foreground">{certificate.fingerprint}</p>
                    </div>
                  </div>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
