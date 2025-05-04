import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Shield, Calendar, Building, Globe } from "lucide-react"

interface CAInfoCardProps {
  className?: string
}

export function CAInfoCard({ className }: CAInfoCardProps) {
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
            <span className="flex h-2 w-2 translate-y-1 rounded-full bg-green-500" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Status: Active</p>
              <p className="text-sm text-muted-foreground">CA is operational and ready to issue certificates</p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Calendar className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Valid Until</p>
              <p className="text-sm text-muted-foreground">January 15, 2035 (10 years remaining)</p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Building className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Organization</p>
              <p className="text-sm text-muted-foreground">LocalCA</p>
            </div>
          </div>
          <div className="grid grid-cols-[25px_1fr] items-start pb-2 last:mb-0 last:pb-0">
            <Globe className="h-5 w-5 text-muted-foreground" />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Common Name</p>
              <p className="text-sm text-muted-foreground">ca.homelab.local</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
