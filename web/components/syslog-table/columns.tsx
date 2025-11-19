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
  id?: number
  timestamp: string
  hostname: string
  facility: number
  severity: number
  priority?: number
  tag: string
  message: string
  pid?: string
  appName?: string
  procID?: string
  msgID?: string
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

const getSeverityConfig = (severity: string, severityNumber: number) => {
  const configs: Record<string, { color: string; textColor: string; bgColor: string; borderColor: string }> = {
    emergency: {
      color: "border-red-600 text-red-700 dark:text-red-300 dark:border-red-500",
      textColor: "text-red-700 dark:text-red-300",
      bgColor: "bg-red-50 dark:bg-red-950/50",
      borderColor: "#dc2626"
    },
    alert: {
      color: "border-orange-600 text-orange-700 dark:text-orange-300 dark:border-orange-500",
      textColor: "text-orange-700 dark:text-orange-300",
      bgColor: "bg-orange-50 dark:bg-orange-950/50",
      borderColor: "#ea580c"
    },
    critical: {
      color: "border-orange-500 text-orange-600 dark:text-orange-300 dark:border-orange-400",
      textColor: "text-orange-600 dark:text-orange-300",
      bgColor: "bg-orange-50 dark:bg-orange-950/50",
      borderColor: "#f97316"
    },
    error: {
      color: "border-red-500 text-red-600 dark:text-red-300 dark:border-red-400",
      textColor: "text-red-600 dark:text-red-300",
      bgColor: "bg-red-50 dark:bg-red-950/50",
      borderColor: "#ef4444"
    },
    warning: {
      color: "border-yellow-500 text-yellow-700 dark:text-yellow-300 dark:border-yellow-500",
      textColor: "text-yellow-700 dark:text-yellow-300",
      bgColor: "bg-yellow-50 dark:bg-yellow-950/50",
      borderColor: "#eab308"
    },
    notice: {
      color: "border-blue-500 text-blue-600 dark:text-blue-300 dark:border-blue-400",
      textColor: "text-blue-600 dark:text-blue-300",
      bgColor: "bg-blue-50 dark:bg-blue-950/50",
      borderColor: "#3b82f6"
    },
    info: {
      color: "border-green-500 text-green-600 dark:text-green-300 dark:border-green-400",
      textColor: "text-green-600 dark:text-green-300",
      bgColor: "bg-green-50 dark:bg-green-950/50",
      borderColor: "#10b981"
    },
    debug: {
      color: "border-gray-400 text-gray-600 dark:text-slate-300 dark:border-slate-500",
      textColor: "text-gray-600 dark:text-slate-300",
      bgColor: "bg-gray-50 dark:bg-slate-800/50",
      borderColor: "#9ca3af"
    },
  }
  return configs[severity] || { color: "border-gray-400 text-gray-600 dark:text-slate-300", textColor: "text-gray-600 dark:text-slate-300", bgColor: "bg-gray-50 dark:bg-slate-800/50", borderColor: "#9ca3af" }
}

// Export for use in data-table for row styling
export { getSeverityConfig, severityNames, facilityNames }

export const columns: ColumnDef<SyslogMessage>[] = [
  {
    accessorKey: "id",
    header: "ID",
    minSize: 60,
    maxSize: 80,
    cell: ({ row }) => {
      const id = row.original.id || row.index + 1
      return <span className="font-mono text-xs text-muted-foreground">{id}</span>
    },
  },
  {
    accessorKey: "timestamp",
    header: "Timestamp",
    minSize: 180,
    maxSize: 200,
    cell: ({ row }) => {
      const timestamp = row.getValue("timestamp") as string
      try {
        const date = new Date(timestamp)
        return (
          <span className="font-mono text-xs">
            {date.toLocaleString('en-US', {
              month: 'short',
              day: '2-digit',
              year: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
              hour12: false
            })}
          </span>
        )
      } catch {
        return <span className="font-mono text-xs">{timestamp}</span>
      }
    },
  },
  {
    accessorKey: "severity",
    header: "Severity",
    minSize: 100,
    maxSize: 120,
    cell: ({ row }) => {
      const severity = row.getValue("severity") as number
      const severityName = severityNames[severity] || "unknown"
      const config = getSeverityConfig(severityName, severity)
      return (
        <span className={`${config.color} ${config.bgColor} border-l-2 px-2 py-0.5 text-xs font-medium capitalize inline-block whitespace-nowrap`}>
          {severityName} {severity}
        </span>
      )
    },
  },
  {
    accessorKey: "hostname",
    header: "Hostname",
    minSize: 120,
    maxSize: 180,
    cell: ({ row }) => (
      <span className="text-xs">{row.getValue("hostname")}</span>
    ),
  },
  {
    accessorKey: "message",
    header: "Message",
    minSize: 300,
    cell: ({ row }) => {
      const message = row.getValue("message") as string
      return (
        <TooltipProvider delayDuration={300}>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="truncate cursor-help text-xs font-mono">
                {message}
              </div>
            </TooltipTrigger>
            <TooltipContent
              side="bottom"
              align="start"
              className="max-w-3xl break-words p-3"
            >
              <p className="whitespace-pre-wrap text-xs font-mono">
                {message}
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )
    },
  },
  {
    accessorKey: "priority",
    header: "Priority",
    minSize: 80,
    maxSize: 90,
    cell: ({ row }) => {
      const facility = row.original.facility
      const severity = row.original.severity
      const priority = row.original.priority || (facility * 8 + severity)
      return <span className="font-mono text-xs text-muted-foreground">{priority}</span>
    },
  },
  {
    accessorKey: "tag",
    header: "App Name",
    minSize: 120,
    maxSize: 180,
    cell: ({ row }) => {
      const tag = row.getValue("tag") as string
      return <span className="text-xs">{tag}</span>
    },
  },
  {
    accessorKey: "pid",
    header: "Proc ID",
    minSize: 70,
    maxSize: 90,
    cell: ({ row }) => {
      const pid = row.original.pid
      return <span className="font-mono text-xs text-muted-foreground">{pid || '-'}</span>
    },
  },
  {
    accessorKey: "msgid",
    header: "Msg ID",
    minSize: 80,
    maxSize: 120,
    cell: ({ row }) => {
      const msgid = row.original.msgid
      return <span className="font-mono text-xs text-muted-foreground">{msgid || '-'}</span>
    },
  },
  {
    accessorKey: "facility",
    header: "Facility",
    minSize: 90,
    maxSize: 120,
    cell: ({ row }) => {
      const facility = row.getValue("facility") as number
      const facilityName = facilityNames[facility] || "unknown"
      return (
        <span className="text-xs">
          {facilityName} {facility}
        </span>
      )
    },
  },
  {
    accessorKey: "structuredData",
    header: "Structured Data",
    minSize: 120,
    maxSize: 200,
    cell: ({ row }) => {
      const data = row.original.structuredData
      return <span className="font-mono text-xs text-muted-foreground truncate">{data || '-'}</span>
    },
  },
]
