import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Plus, RefreshCw, XCircle, Download } from "lucide-react"

interface RecentActivityProps {
  className?: string
}

export function RecentActivity({ className }: RecentActivityProps) {
  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
        <CardDescription>Latest actions performed in the system</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-8">
          <div className="flex items-start">
            <div className="mr-4 mt-0.5">
              <Plus className="h-5 w-5 text-green-500" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Certificate Created</p>
              <p className="text-sm text-muted-foreground">Created server certificate for "api.local"</p>
              <p className="text-xs text-muted-foreground">2 hours ago</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="mr-4 mt-0.5">
              <Download className="h-5 w-5 text-blue-500" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Certificate Downloaded</p>
              <p className="text-sm text-muted-foreground">Downloaded client certificate "john.doe.p12"</p>
              <p className="text-xs text-muted-foreground">5 hours ago</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="mr-4 mt-0.5">
              <RefreshCw className="h-5 w-5 text-amber-500" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Certificate Renewed</p>
              <p className="text-sm text-muted-foreground">Renewed server certificate for "server.local"</p>
              <p className="text-xs text-muted-foreground">Yesterday at 2:30 PM</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="mr-4 mt-0.5">
              <XCircle className="h-5 w-5 text-red-500" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Certificate Revoked</p>
              <p className="text-sm text-muted-foreground">Revoked client certificate "old-client.p12"</p>
              <p className="text-xs text-muted-foreground">2 days ago</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
