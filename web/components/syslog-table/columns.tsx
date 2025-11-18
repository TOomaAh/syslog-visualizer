"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { formatDistanceToNow } from "date-fns"

export interface SyslogMessage {
  timestamp: string
  hostname: string
  facility: number
  severity: number
  tag: string
  message: string
  pid?: string
}

const severityNames: Record<number, string> = {
  0: "emergency",
  1: "alert",
  2: "critical",
  3: "error",
  4: "warning",
  5: "notice",
  6: "info",
  7: "debug",
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

const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    emergency: "bg-red-600 text-white hover:bg-red-700",
    alert: "bg-orange-600 text-white hover:bg-orange-700",
    critical: "bg-orange-500 text-white hover:bg-orange-600",
    error: "bg-red-500 text-white hover:bg-red-600",
    warning: "bg-yellow-500 text-black hover:bg-yellow-600",
    notice: "bg-blue-500 text-white hover:bg-blue-600",
    info: "bg-green-500 text-white hover:bg-green-600",
    debug: "bg-gray-500 text-white hover:bg-gray-600",
  }
  return colors[severity] || "bg-gray-400"
}

export const columns: ColumnDef<SyslogMessage>[] = [
  {
    accessorKey: "timestamp",
    header: "Time",
    size: 200,
    cell: ({ row }) => {
      const timestamp = row.getValue("timestamp") as string
      try {
        const date = new Date(timestamp)
        return (
          <div className="flex flex-col gap-1">
            <span className="font-medium text-sm">
              {formatDistanceToNow(date, { addSuffix: true })}
            </span>
            <span className="font-mono text-xs text-muted-foreground">
              {date.toLocaleString()}
            </span>
          </div>
        )
      } catch {
        return <span className="font-mono text-sm">{timestamp}</span>
      }
    },
  },
  {
    accessorKey: "severity",
    header: "Severity",
    size: 130,
    cell: ({ row }) => {
      const severity = row.getValue("severity") as number
      const severityName = severityNames[severity] || "unknown"
      return (
        <Badge className={`${getSeverityColor(severityName)} px-3 py-1 text-xs font-semibold`}>
          {severityName}
        </Badge>
      )
    },
  },
  {
    accessorKey: "facility",
    header: "Facility",
    size: 110,
    cell: ({ row }) => {
      const facility = row.getValue("facility") as number
      return (
        <span className="text-sm font-medium bg-muted px-2 py-1 rounded">
          {facilityNames[facility] || "unknown"}
        </span>
      )
    },
  },
  {
    accessorKey: "hostname",
    header: "Hostname",
    size: 170,
    cell: ({ row }) => (
      <span className="font-semibold text-sm">{row.getValue("hostname")}</span>
    ),
  },
  {
    accessorKey: "tag",
    header: "Tag",
    size: 170,
    cell: ({ row }) => {
      const tag = row.getValue("tag") as string
      const pid = row.original.pid
      return (
        <div className="flex items-center gap-1.5">
          <span className="text-sm font-medium">{tag}</span>
          {pid && (
            <span className="text-xs text-muted-foreground font-mono bg-muted px-1.5 py-0.5 rounded">
              {pid}
            </span>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: "message",
    header: "Message",
    size: 500,
    cell: ({ row }) => {
      const message = row.getValue("message") as string
      return (
        <TooltipProvider delayDuration={300}>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="truncate cursor-help text-sm leading-relaxed">
                {message}
              </div>
            </TooltipTrigger>
            <TooltipContent
              side="bottom"
              align="start"
              className="max-w-2xl break-words p-4"
            >
              <p className="whitespace-pre-wrap text-sm leading-relaxed">
                {message}
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    },
  },
]
