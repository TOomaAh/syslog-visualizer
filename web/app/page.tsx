"use client"

import { useEffect, useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import { ColumnFiltersState } from "@tanstack/react-table"
import { DataTable } from "@/components/syslog-table/data-table"
import { columns, SyslogMessage } from "@/components/syslog-table/columns"

// Map severity names to numbers
const severityNameToNumber: Record<string, number> = {
  emergency: 0,
  alert: 1,
  critical: 2,
  error: 3,
  warning: 4,
  notice: 5,
  info: 6,
  debug: 7,
}

// Map facility names to numbers
const facilityNameToNumber: Record<string, number> = {
  kern: 0,
  user: 1,
  mail: 2,
  daemon: 3,
  auth: 4,
  syslog: 5,
  lpr: 6,
  news: 7,
  uucp: 8,
  cron: 9,
  authpriv: 10,
  ftp: 11,
  local0: 16,
  local1: 17,
  local2: 18,
  local3: 19,
  local4: 20,
  local5: 21,
  local6: 22,
  local7: 23,
}

interface FilterOptions {
  hostnames: string[]
  tags: string[]
  facilities: number[]
  severities: number[]
}

export default function Home() {
  const [data, setData] = useState<SyslogMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 10 })
  const [totalRows, setTotalRows] = useState(0)
  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null)
  const router = useRouter()

  const fetchData = useCallback(async () => {
    try {
      const params = new URLSearchParams()

      params.append("limit", pagination.pageSize.toString())
      params.append("offset", (pagination.pageIndex * pagination.pageSize).toString())

      const severityFilter = columnFilters.find((f) => f.id === "severity")
      if (severityFilter && Array.isArray(severityFilter.value)) {
        const severities = (severityFilter.value as string[])
          .map((name) => severityNameToNumber[name])
          .filter((num) => num !== undefined)
        if (severities.length > 0) {
          params.append("severities", severities.join(","))
        }
      }

      const facilityFilter = columnFilters.find((f) => f.id === "facility")
      if (facilityFilter && Array.isArray(facilityFilter.value)) {
        const facilities = (facilityFilter.value as string[])
          .map((name) => facilityNameToNumber[name])
          .filter((num) => num !== undefined)
        if (facilities.length > 0) {
          params.append("facilities", facilities.join(","))
        }
      }

      const messageFilter = columnFilters.find((f) => f.id === "message")
      if (messageFilter && typeof messageFilter.value === "string") {
        params.append("search", messageFilter.value)
      }

      const hostnameFilter = columnFilters.find((f) => f.id === "hostname")
      if (hostnameFilter && Array.isArray(hostnameFilter.value)) {
        const hostnames = hostnameFilter.value as string[]
        if (hostnames.length > 0) {
          params.append("hostnames", hostnames.join(","))
        }
      }

      const url = `/api/syslogs?${params.toString()}`
      const response = await fetch(url, {
        credentials: "include",
      })

      if (response.status === 401) {
        router.push("/login")
        return
      }

      if (!response.ok) {
        throw new Error("Failed to fetch syslogs")
      }
      const json = await response.json()

      // API now returns { data: [...], total: N }
      setData(json.data || [])
      setTotalRows(json.total || 0)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred")
    } finally {
      setLoading(false)
    }
  }, [columnFilters, pagination, router])

  useEffect(() => {
    const fetchFilterOptions = async () => {
      try {
        const response = await fetch("/api/filter-options", {
          credentials: "include",
        })

        if (response.status === 401) {
          router.push("/login")
          return
        }

        if (response.ok) {
          const options = await response.json()
          setFilterOptions(options)
        }
      } catch (err) {
        console.error("Failed to fetch filter options:", err)
      }
    }

    fetchFilterOptions()
  }, [router])

  // Reset to first page when filters change
  useEffect(() => {
    setPagination((prev) => ({ ...prev, pageIndex: 0 }))
  }, [columnFilters])

  useEffect(() => {
    fetchData()

    // Poll for new messages every 5 seconds
    const interval = setInterval(fetchData, 5000)

    return () => clearInterval(interval)
  }, [fetchData])

  const handleLogout = async () => {
    try {
      await fetch("/api/auth/logout", {
        method: "POST",
        credentials: "include",
      })
      localStorage.removeItem("apiToken")
      router.push("/login")
    } catch (err) {
      console.error("Logout failed:", err)
    }
  }

  return (
    <main className="min-h-screen p-6 md:p-10 bg-background">
      <div className="mx-auto max-w-[1600px] space-y-8">
        <div className="flex items-center justify-between pb-2">
          <div className="space-y-3">
            <h1 className="text-4xl font-bold tracking-tight">
              Syslog Visualizer
            </h1>
            <p className="text-base text-muted-foreground">
              Real-time syslog monitoring with advanced filtering
            </p>
          </div>
          <button
            onClick={handleLogout}
            className="rounded-lg bg-muted px-5 py-2.5 text-sm font-semibold hover:bg-muted/80 transition-colors"
          >
            Logout
          </button>
        </div>

        {loading && (
          <div className="flex items-center justify-center p-8">
            <div className="text-muted-foreground">Loading...</div>
          </div>
        )}

        {error && (
          <div className="rounded-md bg-destructive/15 p-4 text-destructive">
            Error: {error}
          </div>
        )}

        {!loading && !error && (
          <DataTable
            columns={columns}
            data={data}
            columnFilters={columnFilters}
            onColumnFiltersChange={setColumnFilters}
            pagination={pagination}
            onPaginationChange={setPagination}
            rowCount={totalRows}
            filterOptions={filterOptions}
          />
        )}
      </div>
    </main>
  )
}
