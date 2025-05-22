"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import { AlertCircle, Database, Shield, RefreshCw, Save, TestTube } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { useToast } from "@/hooks/use-toast"

interface SettingsData {
  general: {
    caName: string
    organization: string
    country: string
    tlsEnabled: boolean
  }
  email: {
    emailNotify: boolean
    smtpServer: string
    smtpPort: string
    smtpUser: string
    smtpPassword: string
    smtpUseTLS: boolean
    emailFrom: string
    emailTo: string
  }
  storage: {
    storagePath: string
    backupPath: string
    autoBackup: boolean
  }
  ca: {
    caKeyPassword: string
    crlExpiryDays: string
  }
}

export function SettingsTabs() {
  const { toast } = useToast()
  const [settings, setSettings] = useState<SettingsData>({
    general: {
      caName: "",
      organization: "LocalCA",
      country: "US",
      tlsEnabled: true,
    },
    email: {
      emailNotify: false,
      smtpServer: "",
      smtpPort: "25",
      smtpUser: "",
      smtpPassword: "",
      smtpUseTLS: false,
      emailFrom: "",
      emailTo: "",
    },
    storage: {
      storagePath: "/app/certs",
      backupPath: "",
      autoBackup: false,
    },
    ca: {
      caKeyPassword: "",
      crlExpiryDays: "30",
    },
  })
  const [isLoading, setIsLoading] = useState(false)
  const [isTesting, setIsTesting] = useState(false)

  // Load settings on component mount
  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      const response = await fetch('/api/proxy/settings', {
        credentials: 'include',
      })
      if (response.ok) {
        const data = await response.json()
        if (data.success && data.data) {
          setSettings(data.data)
        }
      }
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  }

  const saveSettings = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/proxy/settings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(settings),
      })

      if (response.ok) {
        const data = await response.json()
        if (data.success) {
          toast({
            title: "Settings saved",
            description: "Your settings have been saved successfully.",
          })
        } else {
          throw new Error(data.message || 'Failed to save settings')
        }
      } else {
        throw new Error('Failed to save settings')
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to save settings",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  const testEmailConfig = async () => {
    setIsTesting(true)
    try {
      const response = await fetch('/api/proxy/test-email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(settings.email),
      })

      if (response.ok) {
        const data = await response.json()
        if (data.success) {
          toast({
            title: "Test email sent",
            description: "Check your inbox for the test email.",
          })
        } else {
          throw new Error(data.message || 'Failed to send test email')
        }
      } else {
        throw new Error('Failed to send test email')
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to send test email",
        variant: "destructive",
      })
    } finally {
      setIsTesting(false)
    }
  }

  const updateSettings = (section: keyof SettingsData, field: string, value: any) => {
    setSettings(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value,
      },
    }))
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
        <Card>
          <CardHeader>
            <CardTitle>General Settings</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="caName">CA Name</Label>
              <Input 
                id="caName" 
                placeholder="e.g., ca.example.com" 
                value={settings.general.caName}
                onChange={(e) => updateSettings('general', 'caName', e.target.value)}
              />
              <p className="text-sm text-muted-foreground">The common name of your Certificate Authority.</p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="organization">Organization</Label>
              <Input 
                id="organization" 
                value={settings.general.organization}
                onChange={(e) => updateSettings('general', 'organization', e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="country">Country</Label>
              <Input 
                id="country" 
                value={settings.general.country}
                onChange={(e) => updateSettings('general', 'country', e.target.value)}
              />
            </div>

            <div className="flex items-center space-x-2">
              <Switch 
                id="tlsEnabled" 
                checked={settings.general.tlsEnabled}
                onCheckedChange={(checked) => updateSettings('general', 'tlsEnabled', checked)}
              />
              <Label htmlFor="tlsEnabled">Enable HTTPS for web interface</Label>
            </div>

            <div className="flex justify-end pt-4">
              <Button onClick={saveSettings} disabled={isLoading} className="flex items-center gap-2">
                <Save className="h-4 w-4" />
                {isLoading ? "Saving..." : "Save Settings"}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="email" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Email Notifications</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch 
                id="emailNotify" 
                checked={settings.email.emailNotify}
                onCheckedChange={(checked) => updateSettings('email', 'emailNotify', checked)}
              />
              <Label htmlFor="emailNotify">Enable email notifications</Label>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="smtpServer">SMTP Server</Label>
              <Input 
                id="smtpServer" 
                placeholder="e.g., smtp.gmail.com" 
                value={settings.email.smtpServer}
                onChange={(e) => updateSettings('email', 'smtpServer', e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="smtpPort">SMTP Port</Label>
              <Input 
                id="smtpPort" 
                placeholder="e.g., 587" 
                value={settings.email.smtpPort}
                onChange={(e) => updateSettings('email', 'smtpPort', e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="smtpUser">SMTP Username</Label>
              <Input 
                id="smtpUser" 
                placeholder="e.g., user@example.com" 
                value={settings.email.smtpUser}
                onChange={(e) => updateSettings('email', 'smtpUser', e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="smtpPassword">SMTP Password</Label>
              <Input 
                id="smtpPassword" 
                type="password" 
                placeholder="Enter SMTP password" 
                value={settings.email.smtpPassword}
                onChange={(e) => updateSettings('email', 'smtpPassword', e.target.value)}
              />
            </div>

            <div className="flex items-center space-x-2">
              <Switch 
                id="smtpUseTLS" 
                checked={settings.email.smtpUseTLS}
                onCheckedChange={(checked) => updateSettings('email', 'smtpUseTLS', checked)}
              />
              <Label htmlFor="smtpUseTLS">Use TLS</Label>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="emailFrom">From Address</Label>
              <Input 
                id="emailFrom" 
                placeholder="e.g., ca@example.com" 
                value={settings.email.emailFrom}
                onChange={(e) => updateSettings('email', 'emailFrom', e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="emailTo">Default Recipient</Label>
              <Input 
                id="emailTo" 
                placeholder="e.g., admin@example.com" 
                value={settings.email.emailTo}
                onChange={(e) => updateSettings('email', 'emailTo', e.target.value)}
              />
            </div>

            <div className="flex justify-between pt-4">
              <Button 
                onClick={testEmailConfig} 
                disabled={isTesting || !settings.email.emailNotify}
                variant="outline"
                className="flex items-center gap-2"
              >
                <TestTube className="h-4 w-4" />
                {isTesting ? "Testing..." : "Test Email Configuration"}
              </Button>
              <Button onClick={saveSettings} disabled={isLoading} className="flex items-center gap-2">
                <Save className="h-4 w-4" />
                {isLoading ? "Saving..." : "Save Settings"}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="storage" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Storage Settings</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="storagePath">Storage Path</Label>
              <Input 
                id="storagePath" 
                value={settings.storage.storagePath}
                onChange={(e) => updateSettings('storage', 'storagePath', e.target.value)}
              />
              <p className="text-sm text-muted-foreground">The directory where certificates and keys are stored.</p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="backupPath">Backup Path</Label>
              <Input 
                id="backupPath" 
                placeholder="e.g., /app/backups" 
                value={settings.storage.backupPath}
                onChange={(e) => updateSettings('storage', 'backupPath', e.target.value)}
              />
              <p className="text-sm text-muted-foreground">Optional: The directory where backups will be stored.</p>
            </div>

            <div className="flex items-center space-x-2">
              <Switch 
                id="autoBackup" 
                checked={settings.storage.autoBackup}
                onCheckedChange={(checked) => updateSettings('storage', 'autoBackup', checked)}
              />
              <Label htmlFor="autoBackup">Enable automatic backups</Label>
            </div>

            <div className="flex justify-between pt-4">
              <Button variant="outline" className="flex items-center gap-2">
                <Database className="h-4 w-4" />
                Backup Now
              </Button>
              <Button onClick={saveSettings} disabled={isLoading} className="flex items-center gap-2">
                <Save className="h-4 w-4" />
                {isLoading ? "Saving..." : "Save Settings"}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="ca" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>CA Management</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
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
                value={settings.ca.caKeyPassword}
                onChange={(e) => updateSettings('ca', 'caKeyPassword', e.target.value)}
              />
              <p className="text-sm text-muted-foreground">Required for CA operations.</p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="crlExpiryDays">CRL Expiry (Days)</Label>
              <Input 
                id="crlExpiryDays" 
                value={settings.ca.crlExpiryDays}
                onChange={(e) => updateSettings('ca', 'crlExpiryDays', e.target.value)}
              />
              <p className="text-sm text-muted-foreground">
                How long the Certificate Revocation List (CRL) is valid.
              </p>
            </div>

            <div className="flex justify-between pt-4">
              <div className="space-x-2">
                <Button variant="outline" className="flex items-center gap-2">
                  <RefreshCw className="h-4 w-4" />
                  Regenerate CRL
                </Button>
                <Button variant="destructive" className="flex items-center gap-2">
                  <Shield className="h-4 w-4" />
                  Regenerate CA Certificate
                </Button>
              </div>
              <Button onClick={saveSettings} disabled={isLoading} className="flex items-center gap-2">
                <Save className="h-4 w-4" />
                {isLoading ? "Saving..." : "Save Settings"}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
