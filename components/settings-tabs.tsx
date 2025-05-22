"use client"

import type React from "react"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import { AlertCircle, Mail, RefreshCw } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { updateSettings, sendTestEmail, regenerateCRL, type Settings } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"

interface SettingsTabsProps {
  initialSettings: Settings
}

export function SettingsTabs({ initialSettings }: SettingsTabsProps) {
  const { toast } = useToast()
  const [isLoading, setIsLoading] = useState({
    general: false,
    email: false,
    storage: false,
    ca: false,
    testEmail: false,
    regenerateCRL: false,
  })

  // Form state
  const [generalSettings, setGeneralSettings] = useState({
    caName: initialSettings.ca_name,
    organization: initialSettings.organization,
    country: initialSettings.country,
    tlsEnabled: initialSettings.tls_enabled,
  })

  const [emailSettings, setEmailSettings] = useState({
    emailNotify: initialSettings.email_notify,
    smtpServer: initialSettings.smtp_server,
    smtpPort: initialSettings.smtp_port,
    smtpUser: initialSettings.smtp_user,
    smtpPassword: initialSettings.smtp_password,
    smtpUseTLS: initialSettings.smtp_use_tls,
    emailFrom: initialSettings.email_from,
    emailTo: initialSettings.email_to,
    testEmailAddress: "",
  })

  const [storageSettings, setStorageSettings] = useState({
    storagePath: initialSettings.storage_path,
    backupPath: initialSettings.backup_path || "",
    autoBackup: initialSettings.auto_backup,
  })

  const [caSettings, setCASettings] = useState({
    caKeyPassword: "",
    crlExpiryDays: initialSettings.crl_expiry_days,
  })

  // Handle form input changes
  const handleGeneralChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value, type, checked } = e.target
    setGeneralSettings((prev) => ({
      ...prev,
      [id]: type === "checkbox" ? checked : value,
    }))
  }

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value, type, checked } = e.target
    setEmailSettings((prev) => ({
      ...prev,
      [id]: type === "checkbox" ? checked : value,
    }))
  }

  const handleStorageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value, type, checked } = e.target
    setStorageSettings((prev) => ({
      ...prev,
      [id]: type === "checkbox" ? checked : value,
    }))
  }

  const handleCAChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value } = e.target
    setCASettings((prev) => ({
      ...prev,
      [id]: value,
    }))
  }

  // Handle form submissions
  const handleGeneralSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setIsLoading((prev) => ({ ...prev, general: true }))

      const response = await updateSettings({
        ca_name: generalSettings.caName,
        organization: generalSettings.organization,
        country: generalSettings.country,
        tls_enabled: generalSettings.tlsEnabled,
      })

      if (response.success) {
        toast({
          title: "Settings Updated",
          description: "General settings have been updated successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to update settings.",
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
      setIsLoading((prev) => ({ ...prev, general: false }))
    }
  }

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setIsLoading((prev) => ({ ...prev, email: true }))

      const response = await updateSettings({
        email_notify: emailSettings.emailNotify,
        smtp_server: emailSettings.smtpServer,
        smtp_port: emailSettings.smtpPort,
        smtp_user: emailSettings.smtpUser,
        smtp_password: emailSettings.smtpPassword,
        smtp_use_tls: emailSettings.smtpUseTLS,
        email_from: emailSettings.emailFrom,
        email_to: emailSettings.emailTo,
      })

      if (response.success) {
        toast({
          title: "Settings Updated",
          description: "Email settings have been updated successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to update settings.",
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
      setIsLoading((prev) => ({ ...prev, email: false }))
    }
  }

  const handleStorageSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setIsLoading((prev) => ({ ...prev, storage: true }))

      const response = await updateSettings({
        storage_path: storageSettings.storagePath,
        backup_path: storageSettings.backupPath,
        auto_backup: storageSettings.autoBackup,
      })

      if (response.success) {
        toast({
          title: "Settings Updated",
          description: "Storage settings have been updated successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to update settings.",
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
      setIsLoading((prev) => ({ ...prev, storage: false }))
    }
  }

  const handleCASubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setIsLoading((prev) => ({ ...prev, ca: true }))

      const response = await updateSettings({
        crl_expiry_days: caSettings.crlExpiryDays,
      })

      if (response.success) {
        toast({
          title: "Settings Updated",
          description: "CA settings have been updated successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to update settings.",
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
      setIsLoading((prev) => ({ ...prev, ca: false }))
    }
  }

  // Handle test email
  const handleTestEmail = async () => {
    if (!emailSettings.testEmailAddress) {
      toast({
        title: "Error",
        description: "Please enter a test email address.",
        variant: "destructive",
      })
      return
    }

    try {
      setIsLoading((prev) => ({ ...prev, testEmail: true }))

      const response = await sendTestEmail(emailSettings.testEmailAddress)

      if (response.success) {
        toast({
          title: "Test Email Sent",
          description: "A test email has been sent successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to send test email.",
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
      setIsLoading((prev) => ({ ...prev, testEmail: false }))
    }
  }

  // Handle CRL regeneration
  const handleRegenerateCRL = async () => {
    try {
      setIsLoading((prev) => ({ ...prev, regenerateCRL: true }))

      const response = await regenerateCRL()

      if (response.success) {
        toast({
          title: "CRL Regenerated",
          description: "The Certificate Revocation List has been regenerated successfully.",
        })
      } else {
        toast({
          title: "Error",
          description: response.message || "Failed to regenerate CRL.",
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
      setIsLoading((prev) => ({ ...prev, regenerateCRL: false }))
    }
  }

  return (
    <Tabs defaultValue="general" className="space-y-4">
      <TabsList>
        <TabsTrigger value="general">General</TabsTrigger>
        <TabsTrigger value="email">Email Notifications</TabsTrigger>
        <TabsTrigger value="storage">Storage</TabsTrigger>
        <TabsTrigger value="ca">CA Management</TabsTrigger>
      </TabsList>
      <TabsContent value="general" className="space-y-4">
        <form onSubmit={handleGeneralSubmit}>
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="caName">CA Name</Label>
                  <Input id="caName" value={generalSettings.caName} onChange={handleGeneralChange} />
                  <p className="text-sm text-muted-foreground">The common name of your Certificate Authority.</p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="organization">Organization</Label>
                  <Input id="organization" value={generalSettings.organization} onChange={handleGeneralChange} />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="country">Country</Label>
                  <Input id="country" value={generalSettings.country} onChange={handleGeneralChange} />
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="tlsEnabled"
                    checked={generalSettings.tlsEnabled}
                    onCheckedChange={(checked) => setGeneralSettings((prev) => ({ ...prev, tlsEnabled: checked }))}
                  />
                  <Label htmlFor="tlsEnabled">Enable HTTPS for web interface</Label>
                </div>

                <Button type="submit" className="mt-4" disabled={isLoading.general}>
                  {isLoading.general ? "Saving..." : "Save General Settings"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </form>
      </TabsContent>
      <TabsContent value="email" className="space-y-4">
        <form onSubmit={handleEmailSubmit}>
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="emailNotify"
                    checked={emailSettings.emailNotify}
                    onCheckedChange={(checked) => setEmailSettings((prev) => ({ ...prev, emailNotify: checked }))}
                  />
                  <Label htmlFor="emailNotify">Enable email notifications</Label>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="smtpServer">SMTP Server</Label>
                  <Input
                    id="smtpServer"
                    placeholder="e.g., smtp.gmail.com"
                    value={emailSettings.smtpServer}
                    onChange={handleEmailChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="smtpPort">SMTP Port</Label>
                  <Input
                    id="smtpPort"
                    placeholder="e.g., 587"
                    value={emailSettings.smtpPort}
                    onChange={handleEmailChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="smtpUser">SMTP Username</Label>
                  <Input
                    id="smtpUser"
                    placeholder="e.g., user@example.com"
                    value={emailSettings.smtpUser}
                    onChange={handleEmailChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="smtpPassword">SMTP Password</Label>
                  <Input
                    id="smtpPassword"
                    type="password"
                    placeholder="Enter SMTP password"
                    value={emailSettings.smtpPassword}
                    onChange={handleEmailChange}
                  />
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="smtpUseTLS"
                    checked={emailSettings.smtpUseTLS}
                    onCheckedChange={(checked) => setEmailSettings((prev) => ({ ...prev, smtpUseTLS: checked }))}
                  />
                  <Label htmlFor="smtpUseTLS">Use TLS</Label>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="emailFrom">From Address</Label>
                  <Input
                    id="emailFrom"
                    placeholder="e.g., ca@example.com"
                    value={emailSettings.emailFrom}
                    onChange={handleEmailChange}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="emailTo">Default Recipient</Label>
                  <Input
                    id="emailTo"
                    placeholder="e.g., admin@example.com"
                    value={emailSettings.emailTo}
                    onChange={handleEmailChange}
                  />
                </div>

                <Button type="submit" className="mt-4" disabled={isLoading.email}>
                  {isLoading.email ? "Saving..." : "Save Email Settings"}
                </Button>

                <div className="grid gap-2 pt-4 border-t">
                  <Label htmlFor="testEmailAddress">Test Email Address</Label>
                  <Input
                    id="testEmailAddress"
                    placeholder="e.g., test@example.com"
                    value={emailSettings.testEmailAddress}
                    onChange={(e) => setEmailSettings((prev) => ({ ...prev, testEmailAddress: e.target.value }))}
                  />

                  <Button
                    type="button"
                    className="flex items-center gap-2 mt-2"
                    onClick={handleTestEmail}
                    disabled={isLoading.testEmail}
                  >
                    <Mail className="h-4 w-4" />
                    {isLoading.testEmail ? "Sending..." : "Test Email Configuration"}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </form>
      </TabsContent>
      <TabsContent value="storage" className="space-y-4">
        <form onSubmit={handleStorageSubmit}>
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="storagePath">Storage Path</Label>
                  <Input id="storagePath" value={storageSettings.storagePath} onChange={handleStorageChange} />
                  <p className="text-sm text-muted-foreground">The directory where certificates and keys are stored.</p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="backupPath">Backup Path</Label>
                  <Input
                    id="backupPath"
                    placeholder="e.g., /app/backups"
                    value={storageSettings.backupPath}
                    onChange={handleStorageChange}
                  />
                  <p className="text-sm text-muted-foreground">Optional: The directory where backups will be stored.</p>
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="autoBackup"
                    checked={storageSettings.autoBackup}
                    onCheckedChange={(checked) => setStorageSettings((prev) => ({ ...prev, autoBackup: checked }))}
                  />
                  <Label htmlFor="autoBackup">Enable automatic backups</Label>
                </div>

                <Button type="submit" className="mt-4" disabled={isLoading.storage}>
                  {isLoading.storage ? "Saving..." : "Save Storage Settings"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </form>
      </TabsContent>
      <TabsContent value="ca" className="space-y-4">
        <form onSubmit={handleCASubmit}>
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
                  <Input
                    id="caKeyPassword"
                    type="password"
                    placeholder="Enter CA key password"
                    value={caSettings.caKeyPassword}
                    onChange={handleCAChange}
                  />
                  <p className="text-sm text-muted-foreground">Required for CA operations.</p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="crlExpiryDays">CRL Expiry (Days)</Label>
                  <Input id="crlExpiryDays" value={caSettings.crlExpiryDays} onChange={handleCAChange} />
                  <p className="text-sm text-muted-foreground">
                    How long the Certificate Revocation List (CRL) is valid.
                  </p>
                </div>

                <Button type="submit" className="mt-4" disabled={isLoading.ca}>
                  {isLoading.ca ? "Saving..." : "Save CA Settings"}
                </Button>

                <div className="grid gap-2 pt-4 border-t">
                  <Button
                    type="button"
                    className="flex items-center gap-2"
                    onClick={handleRegenerateCRL}
                    disabled={isLoading.regenerateCRL}
                  >
                    <RefreshCw className="h-4 w-4" />
                    {isLoading.regenerateCRL ? "Regenerating..." : "Regenerate CRL"}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </form>
      </TabsContent>
    </Tabs>
  )
}
