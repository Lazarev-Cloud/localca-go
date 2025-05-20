"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { AlertCircle, Info, Loader2 } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { useCertificates } from "@/hooks/use-certificates"
import { useToast } from "@/hooks/use-toast"

export function CreateCertificateForm() {
  const [isClientCert, setIsClientCert] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const { createCertificate } = useCertificates()
  const { toast } = useToast()

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setSubmitting(true)

    const formData = new FormData(e.currentTarget)
    
    try {
      const result = await createCertificate(formData)
      
      if (result.success) {
        toast({
          title: "Certificate created",
          description: "The certificate was created successfully.",
        })
        // Reset form
        e.currentTarget.reset()
        setIsClientCert(false)
        setShowAdvanced(false)
      } else {
        toast({
          variant: "destructive",
          title: "Error creating certificate",
          description: result.error || "An unknown error occurred.",
        })
      }
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Error creating certificate",
        description: error instanceof Error ? error.message : "An unknown error occurred.",
      })
    } finally {
      setSubmitting(false)
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
                    name="common_name"
                    placeholder="e.g., server.local or john.doe" 
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
                    name="alt_names"
                    placeholder="e.g., www.server.local, api.server.local"
                    disabled={isClientCert}
                  />
                  <p className="text-sm text-muted-foreground">
                    Additional domain names for this certificate, separated by commas. Only applicable for server
                    certificates.
                  </p>
                </div>

                <div className="flex items-center space-x-2">
                  <Switch 
                    id="isClientCert" 
                    name="is_client"
                    checked={isClientCert} 
                    onCheckedChange={setIsClientCert} 
                  />
                  <Label htmlFor="isClientCert">Create client certificate</Label>
                </div>

                {isClientCert && (
                  <div className="grid gap-2">
                    <Label htmlFor="p12Password">P12 Password</Label>
                    <Input 
                      id="p12Password" 
                      name="p12_password"
                      type="password" 
                      placeholder="Enter a secure password" 
                    />
                    <p className="text-sm text-muted-foreground">
                      This password will be required when importing the certificate into browsers or devices.
                    </p>
                  </div>
                )}

                <div className="grid gap-2">
                  <Label htmlFor="validityPeriod">Validity Period</Label>
                  <Select defaultValue="365" name="validity_days">
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
                    name="organization"
                    placeholder="e.g., Your Company" 
                    defaultValue="LocalCA" 
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="country">Country</Label>
                  <Input 
                    id="country" 
                    name="country"
                    placeholder="e.g., US" 
                    defaultValue="US" 
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="keyType">Key Type</Label>
                  <Select defaultValue="rsa" name="key_type">
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
                  <Select defaultValue="2048" name="key_size">
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
                  <Select defaultValue="sha256" name="signature_algorithm">
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
        <Button variant="outline" type="button" onClick={() => window.history.back()}>Cancel</Button>
        <Button type="submit" disabled={submitting}>
          {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create Certificate
        </Button>
      </div>
    </form>
  )
}
