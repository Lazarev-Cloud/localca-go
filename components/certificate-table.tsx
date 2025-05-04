"use client"

import { useState } from "react"
import Link from "next/link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { MoreHorizontal, Download, RefreshCw, XCircle, AlertTriangle, CheckCircle, FileText } from "lucide-react"

export function CertificateTable() {
  const [certificates] = useState([
    {
      id: "1",
      commonName: "server.local",
      type: "Server",
      expiryDate: "2025-05-01",
      isExpiringSoon: true,
      isExpired: false,
      isRevoked: false,
      serialNumber: "1A:2B:3C:4D:5E:6F",
    },
    {
      id: "2",
      commonName: "api.local",
      type: "Server",
      expiryDate: "2025-08-15",
      isExpiringSoon: false,
      isExpired: false,
      isRevoked: false,
      serialNumber: "7G:8H:9I:10J:11K",
    },
    {
      id: "3",
      commonName: "john.doe",
      type: "Client",
      expiryDate: "2025-06-20",
      isExpiringSoon: false,
      isExpired: false,
      isRevoked: false,
      serialNumber: "12L:13M:14N:15O:16P",
    },
    {
      id: "4",
      commonName: "db.local",
      type: "Server",
      expiryDate: "2025-09-10",
      isExpiringSoon: false,
      isExpired: false,
      isRevoked: false,
      serialNumber: "17Q:18R:19S:20T:21U",
    },
    {
      id: "5",
      commonName: "jane.smith",
      type: "Client",
      expiryDate: "2025-07-05",
      isExpiringSoon: false,
      isExpired: false,
      isRevoked: false,
      serialNumber: "22V:23W:24X:25Y:26Z",
    },
    {
      id: "6",
      commonName: "old-client.p12",
      type: "Client",
      expiryDate: "2025-04-30",
      isExpiringSoon: false,
      isExpired: false,
      isRevoked: true,
      serialNumber: "27A:28B:29C:30D:31E",
    },
    {
      id: "7",
      commonName: "expired.local",
      type: "Server",
      expiryDate: "2023-12-31",
      isExpiringSoon: false,
      isExpired: true,
      isRevoked: false,
      serialNumber: "32F:33G:34H:35I:36J",
    },
  ])

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Common Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Expiry Date</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Serial Number</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {certificates.map((certificate) => (
            <TableRow key={certificate.id}>
              <TableCell className="font-medium">
                <div className="flex items-center space-x-2">
                  <FileText className="h-4 w-4 text-muted-foreground" />
                  <Link href={`/certificates/${certificate.id}`} className="hover:underline">
                    {certificate.commonName}
                  </Link>
                </div>
              </TableCell>
              <TableCell>{certificate.type}</TableCell>
              <TableCell>{certificate.expiryDate}</TableCell>
              <TableCell>
                {certificate.isRevoked ? (
                  <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                    <XCircle className="h-3 w-3" />
                    Revoked
                  </Badge>
                ) : certificate.isExpired ? (
                  <Badge variant="outline" className="flex items-center gap-1 text-red-500 border-red-200 bg-red-50">
                    <XCircle className="h-3 w-3" />
                    Expired
                  </Badge>
                ) : certificate.isExpiringSoon ? (
                  <Badge
                    variant="outline"
                    className="flex items-center gap-1 text-amber-500 border-amber-200 bg-amber-50"
                  >
                    <AlertTriangle className="h-3 w-3" />
                    Expires Soon
                  </Badge>
                ) : (
                  <Badge
                    variant="outline"
                    className="flex items-center gap-1 text-green-500 border-green-200 bg-green-50"
                  >
                    <CheckCircle className="h-3 w-3" />
                    Valid
                  </Badge>
                )}
              </TableCell>
              <TableCell className="font-mono text-xs">{certificate.serialNumber}</TableCell>
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <Download className="mr-2 h-4 w-4" />
                      <span>Download</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <RefreshCw className="mr-2 h-4 w-4" />
                      <span>Renew</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <XCircle className="mr-2 h-4 w-4" />
                      <span>Revoke</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
