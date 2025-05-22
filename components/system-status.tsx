import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { HardDrive, Database, Clock, AlertTriangle } from "lucide-react"
import type { SystemStatus as SystemStatusType } from "@/lib/api"

interface SystemStatusProps {
  className?: string
  systemStatus: SystemStatusType
}

export function SystemStatus({ className, systemStatus }: SystemStatusProps) {
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
                  <div className="text-sm text-muted-foreground">{systemStatus.storage_usage}%</div>
                </div>
                <Progress value={systemStatus.storage_usage} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Database Size</div>
                  </div>
                  <div className="text-sm text-muted-foreground">{systemStatus.database_size}%</div>
                </div>
                <Progress value={systemStatus.database_size} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Uptime</div>
                  </div>
                  <div className="text-sm text-muted-foreground">{systemStatus.uptime}%</div>
                </div>
                <Progress value={systemStatus.uptime} />
              </div>
            </div>
          </TabsContent>
          <TabsContent value="certificates" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{systemStatus.certificate_count}</div>
                  <div className="text-xs text-muted-foreground">Active Certificates</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{systemStatus.expiring_soon_count}</div>
                  <div className="text-xs text-muted-foreground">Expiring Soon</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{systemStatus.revoked_count}</div>
                  <div className="text-xs text-muted-foreground">Revoked</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">{systemStatus.client_certificate_count}</div>
                  <div className="text-xs text-muted-foreground">Client Certificates</div>
                </div>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="alerts" className="space-y-4">
            <div className="grid gap-4 py-4">
              {systemStatus.alerts.length > 0 ? (
                systemStatus.alerts.map((alert, index) => (
                  <div key={index} className="flex items-center gap-4 rounded-lg border p-4">
                    <AlertTriangle className="h-5 w-5 text-amber-500" />
                    <div>
                      <div className="font-medium">{alert.type}</div>
                      <div className="text-sm text-muted-foreground">{alert.message}</div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="flex items-center justify-center p-4 text-muted-foreground">No alerts at this time</div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
