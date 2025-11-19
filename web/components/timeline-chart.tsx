"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"

interface TimelineChartProps {
  filters?: {
    severities?: string
    facilities?: string
    hostnames?: string
  }
}

interface TimeSlot {
  timestamp: string
  severity_counts: Record<string, number>
  total: number
}

const severityColors: Record<number, string> = {
  0: "#dc2626", // emergency - red-600
  1: "#ea580c", // alert - orange-600
  2: "#f97316", // critical - orange-500
  3: "#ef4444", // error - red-500
  4: "#eab308", // warning - yellow-500
  5: "#3b82f6", // notice - blue-500
  6: "#10b981", // info - green-500
  7: "#9ca3af", // debug - gray-400
}

export function TimelineChart({ filters }: TimelineChartProps) {
  const [timeSlots, setTimeSlots] = useState<TimeSlot[]>([])
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    const fetchTimeline = async () => {
      try {
        const params = new URLSearchParams()
        if (filters?.severities) params.append("severities", filters.severities)
        if (filters?.facilities) params.append("facilities", filters.facilities)
        if (filters?.hostnames) params.append("hostnames", filters.hostnames)

        const url = `/api/timeline?${params.toString()}`
        const response = await fetch(url, {
          credentials: "include",
        })

        if (response.status === 401) {
          router.push("/login")
          return
        }

        if (!response.ok) {
          throw new Error("Failed to fetch timeline")
        }

        const data: TimeSlot[] = await response.json()
        setTimeSlots(data)
      } catch (err) {
        console.error("Timeline fetch error:", err)
        // Keep old data on error instead of clearing
      } finally {
        setIsInitialLoading(false)
      }
    }

    fetchTimeline()
  }, [filters, router])

  // Add labels to slots
  const slots = timeSlots.map((slot) => ({
    ...slot,
    label: new Date(slot.timestamp).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
  }))

  const maxCount = Math.max(...slots.map(slot => slot.total), 1)

  if (isInitialLoading) {
    return (
      <div className="w-full bg-white dark:bg-slate-900/50 border dark:border-slate-800 rounded-lg p-3 shadow-md dark:shadow-slate-900/50">
        <div className="h-16 flex items-center justify-center text-xs text-gray-500 dark:text-slate-400">
          Loading timeline...
        </div>
      </div>
    )
  }

  if (slots.length === 0) {
    return null
  }

  return (
    <div className="w-full bg-white dark:bg-slate-900/50 border dark:border-slate-800 rounded-lg p-3 shadow-md dark:shadow-slate-900/50">
      <div className="flex items-end gap-px h-12 mb-4">
        {slots.map((slot, index) => {
          const heightPercent = slot.total > 0 ? Math.max((slot.total / maxCount) * 100, 5) : 0

          // Calculate stacked segments for each severity
          const severityEntries = Object.entries(slot.severity_counts)
            .map(([sev, count]) => ({ severity: Number(sev), count }))
            .sort((a, b) => a.severity - b.severity) // Sort by severity (0=emergency first)

          return (
            <div key={index} className="flex-1 flex flex-col justify-end h-full">
              <div
                className="w-full flex flex-col-reverse rounded-t transition-all hover:opacity-80 cursor-pointer"
                style={{ height: `${heightPercent}%` }}
                title={`${slot.label}: ${slot.total} message${slot.total !== 1 ? 's' : ''}${
                  severityEntries.length > 0
                    ? '\n' + severityEntries.map(e => `Severity ${e.severity}: ${e.count}`).join('\n')
                    : ''
                }`}
              >
                {severityEntries.map((entry, segIndex) => {
                  const segmentHeightPercent = (entry.count / slot.total) * 100
                  return (
                    <div
                      key={segIndex}
                      style={{
                        height: `${segmentHeightPercent}%`,
                        backgroundColor: severityColors[entry.severity],
                      }}
                      className={segIndex === severityEntries.length - 1 ? 'rounded-t' : ''}
                    />
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>
      <div className="flex justify-between text-[9px] text-gray-500 dark:text-slate-400 font-mono px-1">
        <span>{slots[0]?.label}</span>
        {slots.length > 2 && (
          <span>{slots[Math.floor(slots.length / 2)]?.label}</span>
        )}
        <span>{slots[slots.length - 1]?.label}</span>
      </div>
    </div>
  )
}
