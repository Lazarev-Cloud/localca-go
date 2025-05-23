import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { HardDrive, Database, Clock, AlertTriangle } from "lucide-react"
import { useEffect, useState } from "react"
import { useApi } from "@/hooks/use-api"

interface SystemStatusProps {
  className?: string
}

interface SystemStats {
  total_certificates: number
  active_certificates: number
  expiring_soon: number
  expired: number
  revoked: number
  client_certificates: number
  server_certificates: number
  storage: {
    total_size_mb: number
    usage_percentage: number
  }
  uptime_percentage: number
}

interface Certificate {
  common_name: string
  expiry_date: string
  is_client: boolean
  serial_number: string
  is_expired: boolean
  is_expiring_soon: boolean
  is_revoked: boolean
}

export function SystemStatus({ className }: SystemStatusProps) {
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const { fetchApi } = useApi()

  useEffect(() => {
    const fetchStatistics = async () => {
      try {
        const response = await fetchApi<SystemStats>('/statistics')
        if (response.success && response.data) {
          setStats(response.data)
        }
      } catch (error) {
        console.error('Failed to fetch statistics:', error)
      }
    }

    const fetchCertificates = async () => {
      try {
        const response = await fetchApi<{ certificates: Certificate[] }>('/certificates')
        if (response.success && response.data) {
          setCertificates(response.data.certificates || [])
        }
      } catch (error) {
        console.error('Failed to fetch certificates:', error)
      }
    }

    fetchStatistics()
    fetchCertificates()
  }, [fetchApi])

  // Get expiring certificates for alerts
  const expiringCerts = certificates.filter(cert => cert.is_expiring_soon && !cert.is_revoked && !cert.is_expired)
  const storageLimitMB = 1024 // 1GB limit for example

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>System Status</CardTitle>
        <CardDescription>System resources and certificate statistics</CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        <Tabs defaultValue="overview">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="certificates">Certificates</TabsTrigger>
            <TabsTrigger value="alerts">Alerts</TabsTrigger>
          </TabsList>
          <TabsContent value="overview" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <HardDrive className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Storage Usage</div>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {stats?.storage?.usage_percentage ? `${Math.round(stats.storage.usage_percentage)}%` : 'Loading...'}
                  </div>
                </div>
                <Progress value={stats?.storage?.usage_percentage || 0} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Database Size</div>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {stats?.storage?.total_size_mb ? `${Math.round(stats.storage.total_size_mb)} MB` : 'Loading...'}
                  </div>
                </div>
                <Progress value={stats?.storage?.total_size_mb ? (stats.storage.total_size_mb / storageLimitMB) * 100 : 0} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Uptime</div>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {stats?.uptime_percentage ? `${stats.uptime_percentage}%` : 'Loading...'}
                  </div>
                </div>
                <Progress value={stats?.uptime_percentage || 0} />
              </div>
            </div>
          </TabsContent>
          <TabsContent value="certificates" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{stats?.active_certificates || 0}</div>
                  <div className="text-xs text-muted-foreground">Active Certificates</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{stats?.expiring_soon || 0}</div>
                  <div className="text-xs text-muted-foreground">Expiring Soon</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{stats?.revoked || 0}</div>
                  <div className="text-xs text-muted-foreground">Revoked</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{stats?.client_certificates || 0}</div>
                  <div className="text-xs text-muted-foreground">Client Certificates</div>
                </div>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="alerts" className="space-y-4">
            <div className="grid gap-4 py-4">
              {expiringCerts.length > 0 ? (
                expiringCerts.map((cert) => (
                  <div key={cert.serial_number} className="flex items-center gap-4 rounded-lg border p-4">
                    <AlertTriangle className="h-5 w-5 text-amber-500" />
                    <div>
                      <div className="font-medium">Certificate Expiring</div>
                      <div className="text-sm text-muted-foreground">
                        {cert.common_name} expires on {cert.expiry_date}
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="flex items-center gap-4 rounded-lg border p-4">
                  <div className="text-green-500 font-medium">No alerts</div>
                  <div className="text-sm text-muted-foreground">All certificates are valid</div>
                </div>
              )}
              {stats?.storage?.usage_percentage && stats.storage.usage_percentage > 80 && (
                <div className="flex items-center gap-4 rounded-lg border p-4">
                  <AlertTriangle className="h-5 w-5 text-amber-500" />
                  <div>
                    <div className="font-medium">Storage Warning</div>
                    <div className="text-sm text-muted-foreground">
                      Storage usage above 80% ({Math.round(stats.storage.usage_percentage)}%)
                    </div>
                  </div>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
