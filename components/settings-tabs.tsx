"use client"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import { AlertCircle, Mail, Database, Shield, RefreshCw } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

export function SettingsTabs() {
  return (
    <Tabs defaultValue="general" className="space-y-4">
      <TabsList>
        <TabsTrigger value="general">General</TabsTrigger>
        <TabsTrigger value="email">Email Notifications</TabsTrigger>
        <TabsTrigger value="storage">Storage</TabsTrigger>
        <TabsTrigger value="ca">CA Management</TabsTrigger>
      </TabsList>
      <TabsContent value="general" className="space-y-4">
        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="caName">CA Name</Label>
                <Input id="caName" placeholder="e.g., ca.example.com" />
                <p className="text-sm text-muted-foreground">The common name of your Certificate Authority.</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="organization">Organization</Label>
                <Input id="organization" defaultValue="LocalCA" />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="country">Country</Label>
                <Input id="country" defaultValue="US" />
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="tlsEnabled" defaultChecked={true} />
                <Label htmlFor="tlsEnabled">Enable HTTPS for web interface</Label>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="email" className="space-y-4">
        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="flex items-center space-x-2">
                <Switch id="emailNotify" />
                <Label htmlFor="emailNotify">Enable email notifications</Label>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="smtpServer">SMTP Server</Label>
                <Input id="smtpServer" placeholder="e.g., smtp.gmail.com" />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="smtpPort">SMTP Port</Label>
                <Input id="smtpPort" placeholder="e.g., 587" defaultValue="25" />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="smtpUser">SMTP Username</Label>
                <Input id="smtpUser" placeholder="e.g., user@example.com" />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="smtpPassword">SMTP Password</Label>
                <Input id="smtpPassword" type="password" placeholder="Enter SMTP password" />
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="smtpUseTLS" />
                <Label htmlFor="smtpUseTLS">Use TLS</Label>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="emailFrom">From Address</Label>
                <Input id="emailFrom" placeholder="e.g., ca@example.com" />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="emailTo">Default Recipient</Label>
                <Input id="emailTo" placeholder="e.g., admin@example.com" />
              </div>

              <Button className="flex items-center gap-2">
                <Mail className="h-4 w-4" />
                Test Email Configuration
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="storage" className="space-y-4">
        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="storagePath">Storage Path</Label>
                <Input id="storagePath" defaultValue="/app/certs" />
                <p className="text-sm text-muted-foreground">The directory where certificates and keys are stored.</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="backupPath">Backup Path</Label>
                <Input id="backupPath" placeholder="e.g., /app/backups" />
                <p className="text-sm text-muted-foreground">Optional: The directory where backups will be stored.</p>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="autoBackup" />
                <Label htmlFor="autoBackup">Enable automatic backups</Label>
              </div>

              <Button className="flex items-center gap-2">
                <Database className="h-4 w-4" />
                Backup Now
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="ca" className="space-y-4">
        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Warning</AlertTitle>
                <AlertDescription>
                  CA management operations are sensitive and can impact all certificates. Proceed with caution.
                </AlertDescription>
              </Alert>

              <div className="grid gap-2">
                <Label htmlFor="caKeyPassword">CA Key Password</Label>
                <Input id="caKeyPassword" type="password" placeholder="Enter CA key password" />
                <p className="text-sm text-muted-foreground">Required for CA operations.</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="crlExpiryDays">CRL Expiry (Days)</Label>
                <Input id="crlExpiryDays" defaultValue="30" />
                <p className="text-sm text-muted-foreground">
                  How long the Certificate Revocation List (CRL) is valid.
                </p>
              </div>

              <Button className="flex items-center gap-2">
                <RefreshCw className="h-4 w-4" />
                Regenerate CRL
              </Button>

              <Button variant="destructive" className="flex items-center gap-2">
                <Shield className="h-4 w-4" />
                Regenerate CA Certificate
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
