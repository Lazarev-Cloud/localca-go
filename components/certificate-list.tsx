import Link from "next/link"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { FileText, AlertTriangle, CheckCircle } from "lucide-react"

interface CertificateListProps {
  className?: string
}

export function CertificateList({ className }: CertificateListProps) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle>Recent Certificates</CardTitle>
          <CardDescription>Recently created or updated certificates</CardDescription>
        </div>
        <Link href="/certificates" className="text-sm text-muted-foreground hover:underline">
          View all
        </Link>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center justify-between space-x-4">
            <div className="flex items-center space-x-4">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium leading-none">server.local</p>
                <p className="text-sm text-muted-foreground">Server Certificate</p>
              </div>
            </div>
            <div className="flex items-center">
              <Badge variant="outline" className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50">
                <AlertTriangle className="h-3 w-3" />
                Expires Soon
              </Badge>
            </div>
          </div>
          <div className="flex items-center justify-between space-x-4">
            <div className="flex items-center space-x-4">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium leading-none">api.local</p>
                <p className="text-sm text-muted-foreground">Server Certificate</p>
              </div>
            </div>
            <div className="flex items-center">
              <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                <CheckCircle className="h-3 w-3" />
                Valid
              </Badge>
            </div>
          </div>
          <div className="flex items-center justify-between space-x-4">
            <div className="flex items-center space-x-4">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium leading-none">john.doe</p>
                <p className="text-sm text-muted-foreground">Client Certificate</p>
              </div>
            </div>
            <div className="flex items-center">
              <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                <CheckCircle className="h-3 w-3" />
                Valid
              </Badge>
            </div>
          </div>
          <div className="flex items-center justify-between space-x-4">
            <div className="flex items-center space-x-4">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium leading-none">db.local</p>
                <p className="text-sm text-muted-foreground">Server Certificate</p>
              </div>
            </div>
            <div className="flex items-center">
              <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                <CheckCircle className="h-3 w-3" />
                Valid
              </Badge>
            </div>
          </div>
          <div className="flex items-center justify-between space-x-4">
            <div className="flex items-center space-x-4">
              <FileText className="h-6 w-6 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium leading-none">jane.smith</p>
                <p className="text-sm text-muted-foreground">Client Certificate</p>
              </div>
            </div>
            <div className="flex items-center">
              <Badge variant="outline" className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50">
                <CheckCircle className="h-3 w-3" />
                Valid
              </Badge>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
