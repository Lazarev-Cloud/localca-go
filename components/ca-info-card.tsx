import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Shield, Calendar, Building, Globe } from "lucide-react"
import type { CAInfo } from "@/lib/api"

interface CAInfoCardProps {
  className?: string
  caInfo: CAInfo
}

export function CAInfoCard({ className, caInfo }: CAInfoCardProps) {
  // Calculate remaining time in a human-readable format
  const getRemainingTimeText = (days: number) => {
    if (days > 365) {
      const years = Math.floor(days / 365)
      return `${years} ${years === 1 ? "year" : "years"} remaining`
    } else {
      return `${days} ${days === 1 ? "day" : "days"} remaining`
    }
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle>Certificate Authority</CardTitle>
          <CardDescription>CA information and status</CardDescription>
        </div>
        <Shield className="h-5 w-5 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="grid gap-4">
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <span
              className={`flex h-2 w-2 translate-y-1 rounded-full ${caInfo.status === "Active" ? "bg-green-500" : "bg-red-500"}`}
            />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Status: {caInfo.status}</p>
              <p className="text-sm text-muted-foreground">
                {caInfo.status === "Active"
                  ? "CA is operational and ready to issue certificates"
                  : "CA is not operational"}
              </p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Calendar className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Valid Until</p>
              <p className="text-sm text-muted-foreground">
                {caInfo.valid_until} ({getRemainingTimeText(caInfo.remaining_days)})
              </p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Building className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Organization</p>
              <p className="text-sm text-muted-foreground">{caInfo.organization}</p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Globe className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Common Name</p>
              <p className="text-sm text-muted-foreground">{caInfo.common_name}</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
