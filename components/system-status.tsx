import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { HardDrive, Database, Clock, AlertTriangle } from "lucide-react"

interface SystemStatusProps {
  className?: string
}

export function SystemStatus({ className }: SystemStatusProps) {
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
                  <div className="text-sm text-muted-foreground">45%</div>
                </div>
                <Progress value={45} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Database Size</div>
                  </div>
                  <div className="text-sm text-muted-foreground">28%</div>
                </div>
                <Progress value={28} />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <div className="text-sm font-medium">Uptime</div>
                  </div>
                  <div className="text-sm text-muted-foreground">99.9%</div>
                </div>
                <Progress value={99.9} />
              </div>
            </div>
          </TabsContent>
          <TabsContent value="certificates" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">24</div>
                  <div className="text-xs text-muted-foreground">Active Certificates</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">3</div>
                  <div className="text-xs text-muted-foreground">Expiring Soon</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">5</div>
                  <div className="text-xs text-muted-foreground">Revoked</div>
                </div>
                <div className="flex flex-col items-center justify-center rounded-lg border p-4">
                  <div className="text-2xl font-bold">12</div>
                  <div className="text-xs text-muted-foreground">Client Certificates</div>
                </div>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="alerts" className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="flex items-center gap-4 rounded-lg border p-4">
                <AlertTriangle className="h-5 w-5 text-amber-500" />
                <div>
                  <div className="font-medium">Certificate Expiring</div>
                  <div className="text-sm text-muted-foreground">server.local expires in 7 days</div>
                </div>
              </div>
              <div className="flex items-center gap-4 rounded-lg border p-4">
                <AlertTriangle className="h-5 w-5 text-amber-500" />
                <div>
                  <div className="font-medium">Certificate Expiring</div>
                  <div className="text-sm text-muted-foreground">api.local expires in 14 days</div>
                </div>
              </div>
              <div className="flex items-center gap-4 rounded-lg border p-4">
                <AlertTriangle className="h-5 w-5 text-amber-500" />
                <div>
                  <div className="font-medium">Storage Warning</div>
                  <div className="text-sm text-muted-foreground">Storage usage above 40%</div>
                </div>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
