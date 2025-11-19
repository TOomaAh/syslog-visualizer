"use client"

import { Cross2Icon, DownloadIcon } from "@radix-ui/react-icons"
import { Table } from "@tanstack/react-table"
import { AlertCircle, AlertTriangle, Info, XCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { DataTableFacetedFilter } from "./data-table-faceted-filter"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { SyslogMessage } from "./columns"

interface FilterOptions {
  hostnames: string[]
  tags: string[]
  facilities: number[]
  severities: number[]
}

interface DataTableToolbarProps<TData> {
  table: Table<TData>
  filterOptions?: FilterOptions | null
}

export function DataTableToolbar<TData>({
  table,
  filterOptions,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0

  const buildExportUrl = (format: string) => {
    const params = new URLSearchParams()
    params.append('format', format)

    // Get current filters from table state
    const filters = table.getState().columnFilters

    const severityFilter = filters.find(f => f.id === 'severity')
    if (severityFilter && Array.isArray(severityFilter.value)) {
      const severityNameToNumber: Record<string, number> = {
        emergency: 0, alert: 1, critical: 2, error: 3,
        warning: 4, notice: 5, info: 6, debug: 7
      }
      const severities = (severityFilter.value as string[])
        .map(name => severityNameToNumber[name])
        .filter(num => num !== undefined)
      if (severities.length > 0) {
        params.append('severities', severities.join(','))
      }
    }

    const facilityFilter = filters.find(f => f.id === 'facility')
    if (facilityFilter && Array.isArray(facilityFilter.value)) {
      const facilityNameToNumber: Record<string, number> = {
        kern: 0, user: 1, mail: 2, daemon: 3, auth: 4, syslog: 5,
        lpr: 6, news: 7, uucp: 8, cron: 9, authpriv: 10, ftp: 11,
        local0: 16, local1: 17, local2: 18, local3: 19,
        local4: 20, local5: 21, local6: 22, local7: 23
      }
      const facilities = (facilityFilter.value as string[])
        .map(name => facilityNameToNumber[name])
        .filter(num => num !== undefined)
      if (facilities.length > 0) {
        params.append('facilities', facilities.join(','))
      }
    }

    const hostnameFilter = filters.find(f => f.id === 'hostname')
    if (hostnameFilter && Array.isArray(hostnameFilter.value)) {
      const hostnames = hostnameFilter.value as string[]
      if (hostnames.length > 0) {
        params.append('hostnames', hostnames.join(','))
      }
    }

    const messageFilter = filters.find(f => f.id === 'message')
    if (messageFilter && typeof messageFilter.value === 'string') {
      params.append('search', messageFilter.value)
    }

    return `/api/export?${params.toString()}`
  }

  const exportToCSV = () => {
    window.location.href = buildExportUrl('csv')
  }

  const exportToJSON = () => {
    window.location.href = buildExportUrl('json')
  }

  const severityNames: Record<number, { name: string; icon: any }> = {
    0: { name: "emergency", icon: XCircle },
    1: { name: "alert", icon: AlertCircle },
    2: { name: "critical", icon: AlertCircle },
    3: { name: "error", icon: XCircle },
    4: { name: "warning", icon: AlertTriangle },
    5: { name: "notice", icon: Info },
    6: { name: "info", icon: Info },
    7: { name: "debug", icon: Info },
  }

  const facilityNames: Record<number, string> = {
    0: "kern",
    1: "user",
    2: "mail",
    3: "daemon",
    4: "auth",
    5: "syslog",
    6: "lpr",
    7: "news",
    8: "uucp",
    9: "cron",
    10: "authpriv",
    11: "ftp",
    16: "local0",
    17: "local1",
    18: "local2",
    19: "local3",
    20: "local4",
    21: "local5",
    22: "local6",
    23: "local7",
  }

  const severities = filterOptions?.severities
    ? filterOptions.severities.map((sev) => ({
        value: severityNames[sev]?.name || "unknown",
        label:
          severityNames[sev]?.name.charAt(0).toUpperCase() +
          severityNames[sev]?.name.slice(1) || "Unknown",
        icon: severityNames[sev]?.icon || Info,
      }))
    : []

  const facilities = filterOptions?.facilities
    ? filterOptions.facilities.map((fac) => ({
        value: facilityNames[fac] || "unknown",
        label:
          (facilityNames[fac] || "unknown").charAt(0).toUpperCase() +
          (facilityNames[fac] || "unknown").slice(1),
      }))
    : []

  const hostnameOptions = filterOptions?.hostnames
    ? filterOptions.hostnames.map((hostname) => ({
        value: hostname,
        label: hostname,
      }))
    : []

  return (
    <div className="flex items-center justify-between py-2">
      <div className="flex flex-1 items-center space-x-3">
        <Input
          placeholder="Search messages..."
          value={(table.getColumn("message")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("message")?.setFilterValue(event.target.value)
          }
          className="h-10 w-[200px] lg:w-[300px] text-sm"
        />
        {table.getColumn("severity") && (
          <DataTableFacetedFilter
            column={table.getColumn("severity")}
            title="Severity"
            options={severities}
          />
        )}
        {table.getColumn("facility") && (
          <DataTableFacetedFilter
            column={table.getColumn("facility")}
            title="Facility"
            options={facilities}
          />
        )}
        {table.getColumn("hostname") && filterOptions && (
          <DataTableFacetedFilter
            column={table.getColumn("hostname")}
            title="Hostname"
            options={hostnameOptions}
          />
        )}
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => table.resetColumnFilters()}
            className="h-10 px-3 lg:px-4 text-sm"
          >
            Reset
            <Cross2Icon className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="h-10 px-3">
            <DownloadIcon className="mr-2 h-4 w-4" />
            Export
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={exportToCSV}>
            Export CSV
          </DropdownMenuItem>
          <DropdownMenuItem onClick={exportToJSON}>
            Export JSON
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
