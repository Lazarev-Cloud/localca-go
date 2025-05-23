"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Plus, RefreshCw, XCircle, Download, Loader2, AlertTriangle } from "lucide-react"
import { useAuditLogs } from "@/hooks/use-audit-logs"
import { Alert, AlertDescription } from "@/components/ui/alert"

interface RecentActivityProps {
  className?: string
}

interface AuditLog {
  id: number
  action: string
  resource: string
  resource_id?: string
  user_ip?: string
  user_agent?: string
  details?: string
  success: boolean
  error?: string
  created_at: string
}

const getActionIcon = (action: string) => {
  switch (action.toLowerCase()) {
    case 'create':
      return <Plus className="h-5 w-5 text-green-500" />
    case 'download':
      return <Download className="h-5 w-5 text-blue-500" />
    case 'renew':
    case 'update':
      return <RefreshCw className="h-5 w-5 text-amber-500" />
    case 'revoke':
    case 'delete':
      return <XCircle className="h-5 w-5 text-red-500" />
    default:
      return <AlertTriangle className="h-5 w-5 text-gray-500" />
  }
}

const getActionDescription = (log: AuditLog) => {
  const resourceName = log.resource_id || log.resource
  
  switch (log.action.toLowerCase()) {
    case 'create':
      if (log.resource === 'certificate') {
        return `Created ${log.details?.includes('client') ? 'client' : 'server'} certificate for "${resourceName}"`
      }
      return `Created ${log.resource} "${resourceName}"`
    case 'download':
      return `Downloaded ${log.resource} "${resourceName}"`
    case 'renew':
      return `Renewed ${log.resource} certificate for "${resourceName}"`
    case 'revoke':
      return `Revoked ${log.resource} certificate "${resourceName}"`
    case 'delete':
      return `Deleted ${log.resource} "${resourceName}"`
    case 'update':
      return `Updated ${log.resource} settings`
    case 'authenticate':
      return log.success ? 'User logged in successfully' : 'Failed login attempt'
    default:
      return `${log.action} ${log.resource} ${resourceName}`
  }
}

const formatTimeAgo = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (diffInSeconds < 60) {
    return 'Just now'
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60)
    return `${minutes} minute${minutes > 1 ? 's' : ''} ago`
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600)
    return `${hours} hour${hours > 1 ? 's' : ''} ago`
  } else if (diffInSeconds < 604800) {
    const days = Math.floor(diffInSeconds / 86400)
    return `${days} day${days > 1 ? 's' : ''} ago`
  } else {
    return date.toLocaleDateString()
  }
}

export function RecentActivity({ className }: RecentActivityProps) {
  const { auditLogs, loading, error } = useAuditLogs(10, 0)

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
        <CardDescription>Latest actions performed in the system</CardDescription>
      </CardHeader>
      <CardContent>
        {loading && (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {!loading && !error && auditLogs.length === 0 && (
          <div className="py-8 text-center text-muted-foreground">
            No recent activity found.
          </div>
        )}

        {!loading && !error && auditLogs.length > 0 && (
          <div className="space-y-8">
            {auditLogs.slice(0, 5).map((activity) => (
              <div key={activity.id} className="flex items-start">
                <div className="mr-4 mt-0.5">
                  {getActionIcon(activity.action)}
                </div>
                <div className="space-y-1 flex-1">
                  <p className="text-sm font-medium leading-none">
                    {activity.action.charAt(0).toUpperCase() + activity.action.slice(1)} {activity.resource.charAt(0).toUpperCase() + activity.resource.slice(1)}
                    {!activity.success && (
                      <span className="text-red-500 ml-2">(Failed)</span>
                    )}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {getActionDescription(activity)}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatTimeAgo(activity.created_at)}
                  </p>
                  {activity.error && (
                    <p className="text-xs text-red-500">
                      Error: {activity.error}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
