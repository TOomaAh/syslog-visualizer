"use client"

import { useEffect, useState, useCallback, useRef } from "react"
import { useRouter } from "next/navigation"
import { ColumnFiltersState } from "@tanstack/react-table"
import { DataTable } from "@/components/syslog-table/data-table"
import { columns, SyslogMessage } from "@/components/syslog-table/columns"
import { TimelineChart } from "@/components/timeline-chart"

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
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [hasMore, setHasMore] = useState(true)
  const [offset, setOffset] = useState(0)
  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null)
  const [isLiveUpdateEnabled, setIsLiveUpdateEnabled] = useState(true)
  const [isDarkMode, setIsDarkMode] = useState(false)
  const router = useRouter()
  const loadMoreRef = useRef<HTMLDivElement>(null)

  // Load theme preference on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem('theme')
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    const shouldBeDark = savedTheme === 'dark' || (!savedTheme && prefersDark)

    setIsDarkMode(shouldBeDark)
    if (shouldBeDark) {
      document.documentElement.classList.add('dark')
    }
  }, [])

  // Toggle theme
  const toggleTheme = () => {
    setIsDarkMode(!isDarkMode)
    if (!isDarkMode) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  const fetchData = useCallback(async (currentOffset: number, isLoadingMore = false, isSilentRefresh = false) => {
    try {
      if (isLoadingMore) {
        setLoadingMore(true)
      } else if (!isSilentRefresh) {
        setLoading(true)
      }

      const params = new URLSearchParams()
      const pageSize = 50

      params.append("limit", pageSize.toString())
      params.append("offset", currentOffset.toString())

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

      const newData = json.data || []

      if (isLoadingMore) {
        setData(prev => [...prev, ...newData])
      } else {
        setData(newData)
      }

      setOffset(currentOffset + pageSize)
      setHasMore(newData.length === pageSize)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred")
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [columnFilters, router])

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

  // Initial load and filter changes
  useEffect(() => {
    setOffset(0)
    setHasMore(true)
    fetchData(0, false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columnFilters])

  // Poll for new messages every 5 seconds
  useEffect(() => {
    if (!isLiveUpdateEnabled) return

    const interval = setInterval(() => {
      // Only refresh if user hasn't scrolled down (offset is still at initial position)
      if (offset <= 50) {
        fetchData(0, false, true)
      }
    }, 5000)

    return () => clearInterval(interval)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLiveUpdateEnabled, offset])

  // Infinite scroll observer
  useEffect(() => {
    if (!loadMoreRef.current || !hasMore || loadingMore) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          // Load next batch
          fetchData(offset, true)
        }
      },
      { threshold: 0.1 }
    )

    observer.observe(loadMoreRef.current)

    return () => observer.disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hasMore, loadingMore, offset])

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
    <main className="h-screen flex flex-col p-6 md:p-10 bg-background dark:bg-slate-950 overflow-hidden">
      <div className="mx-auto w-full max-w-[95vw] flex flex-col h-full">
        <div className="flex items-center justify-between pb-4 flex-shrink-0">
          <div className="space-y-2">
            <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-slate-100">
              Syslog Visualizer
            </h1>
            <p className="text-sm text-gray-600 dark:text-slate-400">
              Real-time syslog monitoring with advanced filtering
            </p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={toggleTheme}
              className="rounded-lg bg-slate-200 dark:bg-slate-800 px-3 py-2.5 text-sm font-semibold hover:bg-slate-300 dark:hover:bg-slate-700 transition-colors shadow-sm"
              title={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {isDarkMode ? '☀️' : '🌙'}
            </button>
            <button
              onClick={() => setIsLiveUpdateEnabled(!isLiveUpdateEnabled)}
              className={`rounded-lg px-4 py-2.5 text-sm font-semibold transition-colors shadow-sm ${
                isLiveUpdateEnabled
                  ? 'bg-green-100 text-green-700 hover:bg-green-200 dark:bg-green-900/50 dark:text-green-300 dark:hover:bg-green-900/70'
                  : 'bg-slate-200 text-slate-700 hover:bg-slate-300 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700'
              }`}
            >
              {isLiveUpdateEnabled ? '● Live' : '○ Paused'}
            </button>
            <button
              onClick={handleLogout}
              className="rounded-lg bg-slate-200 dark:bg-slate-800 px-5 py-2.5 text-sm font-semibold hover:bg-slate-300 dark:hover:bg-slate-700 transition-colors shadow-sm"
            >
              Logout
            </button>
          </div>
        </div>

        {loading && (
          <div className="flex items-center justify-center p-8 flex-grow">
            <div className="text-gray-600 dark:text-slate-400">Loading...</div>
          </div>
        )}

        {error && (
          <div className="rounded-lg bg-red-50 dark:bg-red-950/50 p-4 text-red-600 dark:text-red-400 flex-shrink-0 border border-red-200 dark:border-red-800/50 shadow-sm">
            Error: {error}
          </div>
        )}

        {!loading && !error && (
          <div className="flex flex-col gap-4 flex-grow min-h-0">
            <TimelineChart filters={{
              severities: columnFilters.find(f => f.id === "severity")
                ? (columnFilters.find(f => f.id === "severity")?.value as string[])
                    .map(name => severityNameToNumber[name])
                    .filter(num => num !== undefined)
                    .join(",")
                : undefined,
              facilities: columnFilters.find(f => f.id === "facility")
                ? (columnFilters.find(f => f.id === "facility")?.value as string[])
                    .map(name => facilityNameToNumber[name])
                    .filter(num => num !== undefined)
                    .join(",")
                : undefined,
              hostnames: columnFilters.find(f => f.id === "hostname")
                ? (columnFilters.find(f => f.id === "hostname")?.value as string[]).join(",")
                : undefined,
            }} />
            <DataTable
              columns={columns}
              data={data}
              columnFilters={columnFilters}
              onColumnFiltersChange={setColumnFilters}
              filterOptions={filterOptions}
              loadMoreRef={loadMoreRef}
              loadingMore={loadingMore}
              hasMore={hasMore}
            />
          </div>
        )}
      </div>
    </main>
  )
}
