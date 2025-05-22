import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Plus, RefreshCw, XCircle, Download } from "lucide-react"
import type { RecentActivity as RecentActivityType } from "@/lib/api"

interface RecentActivityProps {
  className?: string
  activities: RecentActivityType[]
}

export function RecentActivity({ className, activities }: RecentActivityProps) {
  // Function to get the appropriate icon based on activity type
  const getActivityIcon = (type: string) => {
    switch (type) {
      case "create":
        return <Plus className="h-5 w-5 text-green-500" />
      case "download":
        return <Download className="h-5 w-5 text-blue-500" />
      case "renew":
        return <RefreshCw className="h-5 w-5 text-amber-500" />
      case "revoke":
        return <XCircle className="h-5 w-5 text-red-500" />
      default:
        return <Plus className="h-5 w-5 text-muted-foreground" />
    }
  }

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
        <CardDescription>Latest actions performed in the system</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-8">
          {activities.length > 0 ? (
            activities.map((activity) => (
              <div key={activity.id} className="flex items-start">
                <div className="mr-4 mt-0.5">{getActivityIcon(activity.type)}</div>
                <div className="space-y-1">
                  <p className="text-sm font-medium leading-none">
                    {activity.type.charAt(0).toUpperCase() + activity.type.slice(1)}
                  </p>
                  <p className="text-sm text-muted-foreground">{activity.message}</p>
                  <p className="text-xs text-muted-foreground">{activity.timestamp}</p>
                </div>
              </div>
            ))
          ) : (
            <div className="flex items-center justify-center p-4 text-muted-foreground">No recent activity</div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
