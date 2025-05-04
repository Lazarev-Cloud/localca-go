import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Download, RefreshCw, XCircle, MoreHorizontal } from "lucide-react"

interface CertificateActionsProps {
  id: string
}

export function CertificateActions({ id }: CertificateActionsProps) {
  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" className="flex items-center gap-2">
        <Download className="h-4 w-4" />
        Download
      </Button>
      <Button variant="outline" className="flex items-center gap-2">
        <RefreshCw className="h-4 w-4" />
        Renew
      </Button>
      <Button variant="outline" className="flex items-center gap-2 text-red-500 hover:text-red-600">
        <XCircle className="h-4 w-4" />
        Revoke
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>More Actions</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>Download as PEM</DropdownMenuItem>
          <DropdownMenuItem>Download as P12</DropdownMenuItem>
          <DropdownMenuItem>Download Private Key</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem>View Certificate Chain</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
