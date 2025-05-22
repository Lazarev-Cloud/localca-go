"use client"

import type React from "react"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { AlertCircle, Info } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { createCertificate } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"

export function CreateCertificateForm() {
  const [isClientCert, setIsClientCert] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()
  const { toast } = useToast()

  // Form state
  const [formData, setFormData] = useState({
    commonName: "",
    alternativeNames: "",
    p12Password: "",
    validityPeriod: "365",
    organization: "LocalCA",
    country: "US",
    keyType: "rsa",
    keySize: "2048",
    signatureAlgorithm: "sha256",
  })

  // Handle form input changes
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { id, value } = e.target
    setFormData((prev) => ({ ...prev, [id]: value }))
  }

  // Handle select changes
  const handleSelectChange = (id: string, value: string) => {
    setFormData((prev) => ({ ...prev, [id]: value }))
  }

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      setIsLoading(true)

      // Prepare data for API
      const certificateData = {
        common_name: formData.commonName,
        type: isClientCert ? "Client" : "Server",
        validity_days: Number.parseInt(formData.validityPeriod),
        organization: formData.organization,
        country: formData.country,
        key_type: formData.keyType,
        key_size: formData.keySize,
        signature_algorithm: formData.signatureAlgorithm,
      }

      // Add client-specific or server-specific fields
      if (isClientCert) {
        if (formData.p12Password) {
          Object.assign(certificateData, { p12_password: formData.p12Password })
        }
      } else {
        if (formData.alternativeNames) {
          const sans = formData.alternativeNames
            .split(",")
            .map((name) => name.trim())
            .filter((name) => name.length > 0)

          if (sans.length > 0) {
            Object.assign(certificateData, { alternative_names: sans })
          }
        }
      }

      const response = await createCertificate(certificateData)

      if (response.success) {
        toast({
          title: "Certificate Created",
          description: "The certificate has been successfully created.",
        })

        // Redirect to certificates page
        router.push("/certificates")
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to create certificate.",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "An unknown error occurred",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form className="space-y-8" onSubmit={handleSubmit}>
      <Tabs defaultValue="basic" className="space-y-4">
        <TabsList>
          <TabsTrigger value="basic">Basic Information</TabsTrigger>
          <TabsTrigger value="advanced" onClick={() => setShowAdvanced(true)}>
            Advanced Options
          </TabsTrigger>
        </TabsList>
        <TabsContent value="basic" className="space-y-4">
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="commonName">Common Name</Label>
                  <Input
                    id="commonName"
                    placeholder="e.g., server.local or john.doe"
                    value={formData.commonName}
                    onChange={handleInputChange}
                    required
                  />
                  <p className="text-sm text-muted-foreground">
                    For server certificates, use the server's hostname. For client certificates, use the user's name.
                  </p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="alternativeNames">Subject Alternative Names (SAN)</Label>
                  <Textarea
                    id="alternativeNames"
                    placeholder="e.g., www.server.local, api.server.local"
                    disabled={isClientCert}
                    value={formData.alternativeNames}
                    onChange={handleInputChange}
                  />
                  <p className="text-sm text-muted-foreground">
                    Additional domain names for this certificate, separated by commas. Only applicable for server
                    certificates.
                  </p>
                </div>

                <div className="flex items-center space-x-2">
                  <Switch id="isClientCert" checked={isClientCert} onCheckedChange={setIsClientCert} />
                  <Label htmlFor="isClientCert">Create client certificate</Label>
                </div>

                {isClientCert && (
                  <div className="grid gap-2">
                    <Label htmlFor="p12Password">P12 Password</Label>
                    <Input
                      id="p12Password"
                      type="password"
                      placeholder="Enter a secure password"
                      value={formData.p12Password}
                      onChange={handleInputChange}
                    />
                    <p className="text-sm text-muted-foreground">
                      This password will be required when importing the certificate into browsers or devices.
                    </p>
                  </div>
                )}

                <div className="grid gap-2">
                  <Label htmlFor="validityPeriod">Validity Period</Label>
                  <Select
                    defaultValue={formData.validityPeriod}
                    onValueChange={(value) => handleSelectChange("validityPeriod", value)}
                  >
                    <SelectTrigger id="validityPeriod">
                      <SelectValue placeholder="Select validity period" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="30">30 days</SelectItem>
                      <SelectItem value="90">90 days</SelectItem>
                      <SelectItem value="180">180 days</SelectItem>
                      <SelectItem value="365">1 year</SelectItem>
                      <SelectItem value="730">2 years</SelectItem>
                      <SelectItem value="1095">3 years</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-muted-foreground">
                    How long the certificate will be valid. Shorter periods are more secure.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Important</AlertTitle>
            <AlertDescription>
              {isClientCert
                ? "Client certificates are used for user authentication. Make sure to securely distribute the P12 file to the intended user."
                : "Server certificates should be installed on your server. The private key should be kept secure and not shared."}
            </AlertDescription>
          </Alert>
        </TabsContent>

        <TabsContent value="advanced" className="space-y-4">
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="organization">Organization</Label>
                  <Input
                    id="organization"
                    placeholder="e.g., Your Company"
                    defaultValue={formData.organization}
                    onChange={handleInputChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="country">Country</Label>
                  <Input
                    id="country"
                    placeholder="e.g., US"
                    defaultValue={formData.country}
                    onChange={handleInputChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="keyType">Key Type</Label>
                  <Select
                    defaultValue={formData.keyType}
                    onValueChange={(value) => handleSelectChange("keyType", value)}
                  >
                    <SelectTrigger id="keyType">
                      <SelectValue placeholder="Select key type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="rsa">RSA</SelectItem>
                      <SelectItem value="ecdsa">ECDSA</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="keySize">Key Size</Label>
                  <Select
                    defaultValue={formData.keySize}
                    onValueChange={(value) => handleSelectChange("keySize", value)}
                  >
                    <SelectTrigger id="keySize">
                      <SelectValue placeholder="Select key size" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="2048">2048 bits (RSA)</SelectItem>
                      <SelectItem value="4096">4096 bits (RSA)</SelectItem>
                      <SelectItem value="256">P-256 (ECDSA)</SelectItem>
                      <SelectItem value="384">P-384 (ECDSA)</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-muted-foreground">
                    Larger key sizes are more secure but may impact performance.
                  </p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="signatureAlgorithm">Signature Algorithm</Label>
                  <Select
                    defaultValue={formData.signatureAlgorithm}
                    onValueChange={(value) => handleSelectChange("signatureAlgorithm", value)}
                  >
                    <SelectTrigger id="signatureAlgorithm">
                      <SelectValue placeholder="Select signature algorithm" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="sha256">SHA-256</SelectItem>
                      <SelectItem value="sha384">SHA-384</SelectItem>
                      <SelectItem value="sha512">SHA-512</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>Advanced Settings</AlertTitle>
            <AlertDescription>
              These settings are optional. The default values are recommended for most use cases.
            </AlertDescription>
          </Alert>
        </TabsContent>
      </Tabs>

      <div className="flex justify-end space-x-4">
        <Button type="button" variant="outline" onClick={() => router.push("/certificates")} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? "Creating..." : "Create Certificate"}
        </Button>
      </div>
    </form>
  )
}
