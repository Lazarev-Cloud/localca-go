import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AlertTriangle, CheckCircle, XCircle, Calendar, FileText, Key, Globe, Building, Hash } from "lucide-react"

interface CertificateDetailsProps {
  id: string
}

export function CertificateDetails({ id }: CertificateDetailsProps) {
  // In a real application, this would fetch the certificate details from the API
  const certificate = {
    id,
    commonName: "server.local",
    type: "Server",
    expiryDate: "2025-05-01",
    issuedDate: "2023-05-01",
    isExpiringSoon: true,
    isExpired: false,
    isRevoked: false,
    serialNumber: "1A:2B:3C:4D:5E:6F",
    organization: "LocalCA",
    country: "US",
    alternativeNames: ["www.server.local", "api.server.local"],
    keyUsage: ["Digital Signature", "Key Encipherment"],
    extendedKeyUsage: ["Server Authentication", "Client Authentication"],
    keyType: "RSA",
    keySize: "2048 bits",
    signatureAlgorithm: "SHA256withRSA",
    fingerprint: "12:34:56:78:9A:BC:DE:F0:12:34:56:78:9A:BC:DE:F0",
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="space-y-1">
            <CardTitle className="text-2xl">{certificate.commonName}</CardTitle>
            <CardDescription>{certificate.type} Certificate</CardDescription>
          </div>
          <div>
            {certificate.isRevoked ? (
              <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                <XCircle className="h-3 w-3" />
                Revoked
              </Badge>
            ) : certificate.isExpired ? (
              <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                <XCircle className="h-3 w-3" />
                Expired
              </Badge>
            ) : certificate.isExpiringSoon ? (
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
                    <p className="text-sm text-muted-foreground">{certificate.commonName}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Valid From</p>
                    <p className="text-sm text-muted-foreground">{certificate.issuedDate}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Valid Until</p>
                    <p className="text-sm text-muted-foreground">{certificate.expiryDate}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Hash className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Serial Number</p>
                    <p className="text-sm font-mono text-muted-foreground">{certificate.serialNumber}</p>
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
                    <p className="text-sm text-muted-foreground">{certificate.organization}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Globe className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Country</p>
                    <p className="text-sm text-muted-foreground">{certificate.country}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Globe className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Alternative Names</p>
                    <div className="flex flex-wrap gap-2 mt-1">
                      {certificate.alternativeNames.map((name, index) => (
                        <Badge key={index} variant="secondary">
                          {name}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Key className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Key Information</p>
                    <p className="text-sm text-muted-foreground">
                      {certificate.keyType}, {certificate.keySize}
                    </p>
                  </div>
                </div>
              </div>
            </TabsContent>
            <TabsContent value="extensions" className="space-y-4">
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Key className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Key Usage</p>
                    <div className="flex flex-wrap gap-2 mt-1">
                      {certificate.keyUsage.map((usage, index) => (
                        <Badge key={index} variant="secondary">
                          {usage}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Key className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Extended Key Usage</p>
                    <div className="flex flex-wrap gap-2 mt-1">
                      {certificate.extendedKeyUsage.map((usage, index) => (
                        <Badge key={index} variant="secondary">
                          {usage}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Hash className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Signature Algorithm</p>
                    <p className="text-sm text-muted-foreground">{certificate.signatureAlgorithm}</p>
                  </div>
                </div>
                <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
                  <Hash className="h-5 w-5 text-muted-foreground" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">Fingerprint (SHA-1)</p>
                    <p className="text-sm font-mono text-muted-foreground">{certificate.fingerprint}</p>
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
