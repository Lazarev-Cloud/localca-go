"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { HardDrive, Database, Clock, AlertTriangle, CheckCircle, TrendingUp } from "lucide-react"
import { useEffect, useState, useCallback } from "react"
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
  const [loading, setLoading] = useState(true)
  const [lastFetch, setLastFetch] = useState<number>(0)
  const { fetchApi } = useApi()

  const fetchData = useCallback(async () => {
    // Debounce: Don't fetch if we've fetched within the last 10 seconds
    const now = Date.now()
    if (now - lastFetch < 10000) {
      return
    }

    try {
      setLoading(true)
      setLastFetch(now)
      
      // Fetch statistics and certificates in parallel
      const [statsResponse, certsResponse] = await Promise.all([
        fetchApi<SystemStats>('/statistics'),
        fetchApi<{ certificates: Certificate[] }>('/certificates')
      ])
      
      if (statsResponse.success && statsResponse.data) {
        setStats(statsResponse.data)
      }
      
      if (certsResponse.success && certsResponse.data) {
        setCertificates(certsResponse.data.certificates || [])
      }
    } catch (error) {
      console.error('Failed to fetch system data:', error)
    } finally {
      setLoading(false)
    }
  }, [fetchApi, lastFetch])

  useEffect(() => {
    fetchData()
    
    // Refresh data every 60 seconds instead of 30 to reduce load
    const interval = setInterval(fetchData, 60000)
    return () => clearInterval(interval)
  }, []) // Remove fetchApi from dependencies to prevent infinite loop

  // Get expiring certificates for alerts
  const expiringCerts = certificates.filter(cert => cert.is_expiring_soon && !cert.is_revoked && !cert.is_expired)
  
  // Calculate storage limit based on usage (assume 1GB default, but calculate based on current usage)
  const calculateStorageLimit = () => {
    if (!stats?.storage?.usage_percentage || !stats?.storage?.total_size_mb) return 1024
    
    // If usage is very low, assume a smaller limit for better visualization
    if (stats.storage.usage_percentage < 1) {
      return Math.max(100, stats.storage.total_size_mb * 10) // Assume 10x current usage as limit
    }
    
    // Calculate limit based on current usage percentage
    return (stats.storage.total_size_mb / stats.storage.usage_percentage) * 100
  }
  
  const storageLimitMB = calculateStorageLimit()

  // Format file size
  const formatFileSize = (mb: number) => {
    if (mb < 1) {
      return `${Math.round(mb * 1024)} KB`
    } else if (mb < 1024) {
      return `${Math.round(mb * 10) / 10} MB`
    } else {
      return `${Math.round((mb / 1024) * 10) / 10} GB`
    }
  }

  // Get status color based on percentage
  const getStatusColor = (percentage: number, thresholds = { warning: 70, critical: 90 }) => {
    if (percentage >= thresholds.critical) return "text-red-500"
    if (percentage >= thresholds.warning) return "text-amber-500"
    return "text-green-500"
  }

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="h-5 w-5" />
          System Status
        </CardTitle>
        <CardDescription>System resources and certificate statistics</CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        <Tabs defaultValue="overview">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="certificates">Certificates</TabsTrigger>
            <TabsTrigger value="alerts">Alerts</TabsTrigger>
          </TabsList>
          
          <TabsContent value="overview" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid gap-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <HardDrive className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Storage Usage</div>
                  </div>
                  <div className={`text-sm font-medium ${getStatusColor(stats?.storage?.usage_percentage || 0)}`}>
                    {loading ? 'Loading...' : `${Math.round(stats?.storage?.usage_percentage || 0)}%`}
                  </div>
                </div>
                <Progress value={stats?.storage?.usage_percentage || 0} className="h-2" />
                <div className="text-xs text-muted-foreground">
                  {loading ? 'Calculating...' : `${formatFileSize(stats?.storage?.total_size_mb || 0)} of ${formatFileSize(storageLimitMB)} used`}
                </div>
              </div>
              
              <div className="grid gap-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Database Size</div>
                  </div>
                  <div className="text-sm font-medium text-blue-600">
                    {loading ? 'Loading...' : formatFileSize(stats?.storage?.total_size_mb || 0)}
                  </div>
                </div>
                <Progress value={stats?.storage?.total_size_mb ? Math.min((stats.storage.total_size_mb / storageLimitMB) * 100, 100) : 0} className="h-2" />
                <div className="text-xs text-muted-foreground">
                  {loading ? 'Calculating...' : `Total data stored: ${formatFileSize(stats?.storage?.total_size_mb || 0)}`}
                </div>
              </div>
              
              <div className="grid gap-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">System Uptime</div>
                  </div>
                  <div className={`text-sm font-medium ${getStatusColor(stats?.uptime_percentage || 0, { warning: 95, critical: 90 })}`}>
                    {loading ? 'Loading...' : `${stats?.uptime_percentage || 0}%`}
                  </div>
                </div>
                <Progress value={stats?.uptime_percentage || 0} className="h-2" />
                <div className="text-xs text-muted-foreground">
                  {loading ? 'Calculating...' : 'Service availability over time'}
                </div>
              </div>
            </div>
          </TabsContent>
          
          <TabsContent value="certificates" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col items-center justify-center rounded-lg border p-4 bg-green-50 border-green-200">
                  <div className="text-2xl font-bold text-green-700">{loading ? '...' : stats?.active_certificates || 0}</div>
                  <div className="text-xs text-green-600 font-medium">Active Certificates</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4 bg-amber-50 border-amber-200">
                  <div className="text-2xl font-bold text-amber-700">{loading ? '...' : stats?.expiring_soon || 0}</div>
                  <div className="text-xs text-amber-600 font-medium">Expiring Soon</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4 bg-red-50 border-red-200">
                  <div className="text-2xl font-bold text-red-700">{loading ? '...' : stats?.revoked || 0}</div>
                  <div className="text-xs text-red-600 font-medium">Revoked</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4 bg-blue-50 border-blue-200">
                  <div className="text-2xl font-bold text-blue-700">{loading ? '...' : stats?.client_certificates || 0}</div>
                  <div className="text-xs text-blue-600 font-medium">Client Certificates</div>
                </div>
              </div>
              
              <div className="mt-4 p-4 rounded-lg bg-gray-50 border">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Total Certificates</span>
                  <span className="text-lg font-bold">{loading ? '...' : stats?.total_certificates || 0}</span>
                </div>
                <div className="flex items-center justify-between mt-2">
                  <span className="text-sm text-muted-foreground">Server Certificates</span>
                  <span className="text-sm font-medium">{loading ? '...' : stats?.server_certificates || 0}</span>
                </div>
              </div>
            </div>
          </TabsContent>
          
          <TabsContent value="alerts" className="space-y-4">
            <div className="grid gap-4 py-4">
              {loading ? (
                <div className="flex items-center gap-4 rounded-lg border p-4">
                  <div className="text-muted-foreground">Loading alerts...</div>
                </div>
              ) : expiringCerts.length > 0 ? (
                expiringCerts.map((cert) => (
                  <div key={cert.serial_number} className="flex items-center gap-4 rounded-lg border p-4 bg-amber-50 border-amber-200">
                    <AlertTriangle className="h-5 w-5 text-amber-500 flex-shrink-0" />
                    <div className="flex-1">
                      <div className="font-medium text-amber-800">Certificate Expiring Soon</div>
                      <div className="text-sm text-amber-700">
                        "{cert.common_name}" expires on {new Date(cert.expiry_date).toLocaleDateString()}
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="flex items-center gap-4 rounded-lg border p-4 bg-green-50 border-green-200">
                  <CheckCircle className="h-5 w-5 text-green-500" />
                  <div>
                    <div className="text-green-800 font-medium">No Certificate Alerts</div>
                    <div className="text-sm text-green-700">All certificates are valid and not expiring soon</div>
                  </div>
                </div>
              )}
              
              {!loading && stats?.storage?.usage_percentage && stats.storage.usage_percentage > 80 && (
                <div className="flex items-center gap-4 rounded-lg border p-4 bg-amber-50 border-amber-200">
                  <AlertTriangle className="h-5 w-5 text-amber-500 flex-shrink-0" />
                  <div>
                    <div className="font-medium text-amber-800">Storage Warning</div>
                    <div className="text-sm text-amber-700">
                      Storage usage above 80% ({Math.round(stats.storage.usage_percentage)}% - {formatFileSize(stats.storage.total_size_mb)})
                    </div>
                  </div>
                </div>
              )}
              
              {!loading && stats?.uptime_percentage && stats.uptime_percentage < 95 && (
                <div className="flex items-center gap-4 rounded-lg border p-4 bg-red-50 border-red-200">
                  <AlertTriangle className="h-5 w-5 text-red-500 flex-shrink-0" />
                  <div>
                    <div className="font-medium text-red-800">Uptime Alert</div>
                    <div className="text-sm text-red-700">
                      System uptime below 95% ({stats.uptime_percentage}%)
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

