"use client"

import { createContext, useContext, useState, ReactNode } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Search, Filter, X } from "lucide-react"

interface CertificateFilters {
  searchQuery: string
  certificateType: string
  status: string
}

interface CertificateFiltersContextType {
  filters: CertificateFilters
  setFilters: (filters: CertificateFilters) => void
  clearFilters: () => void
}

const CertificateFiltersContext = createContext<CertificateFiltersContextType | undefined>(undefined)

export function CertificateFiltersProvider({ children }: { children: ReactNode }) {
  const [filters, setFilters] = useState<CertificateFilters>({
    searchQuery: "",
    certificateType: "all",
    status: "all"
  })

  const clearFilters = () => {
    setFilters({
      searchQuery: "",
      certificateType: "all",
      status: "all"
    })
  }

  return (
    <CertificateFiltersContext.Provider value={{ filters, setFilters, clearFilters }}>
      {children}
    </CertificateFiltersContext.Provider>
  )
}

export function useCertificateFilters() {
  const context = useContext(CertificateFiltersContext)
  if (context === undefined) {
    throw new Error('useCertificateFilters must be used within a CertificateFiltersProvider')
  }
  return context
}

export function CertificateFilters() {
  const { filters, setFilters, clearFilters } = useCertificateFilters()

  const updateFilter = (key: keyof CertificateFilters, value: string) => {
    setFilters({ ...filters, [key]: value })
  }

  const hasActiveFilters = filters.searchQuery !== "" || filters.certificateType !== "all" || filters.status !== "all"

  return (
    <div className="flex flex-col space-y-4 md:flex-row md:items-end md:space-x-4 md:space-y-0">
      <div className="flex-1 space-y-2">
        <div className="text-sm font-medium">Search</div>
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search by common name or serial number..."
            className="pl-8"
            value={filters.searchQuery}
            onChange={(e) => updateFilter('searchQuery', e.target.value)}
          />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4 md:flex md:flex-row md:space-x-4">
        <div className="space-y-2">
          <div className="text-sm font-medium">Type</div>
          <Select value={filters.certificateType} onValueChange={(value) => updateFilter('certificateType', value)}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="All Types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Types</SelectItem>
              <SelectItem value="server">Server</SelectItem>
              <SelectItem value="client">Client</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <div className="text-sm font-medium">Status</div>
          <Select value={filters.status} onValueChange={(value) => updateFilter('status', value)}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="All Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="valid">Valid</SelectItem>
              <SelectItem value="expiring">Expiring Soon</SelectItem>
              <SelectItem value="expired">Expired</SelectItem>
              <SelectItem value="revoked">Revoked</SelectItem>
            </SelectContent>
          </Select>
        </div>
        {hasActiveFilters && (
          <Button variant="outline" onClick={clearFilters} className="flex items-center gap-2">
            <X className="h-4 w-4" />
            Clear Filters
          </Button>
        )}
      </div>
    </div>
  )
}
