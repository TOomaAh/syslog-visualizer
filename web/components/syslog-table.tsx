"use client";

import * as React from "react";
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";

export interface SyslogMessage {
  id: string;
  timestamp: string;
  hostname: string;
  facility: string;
  severity: string;
  tag: string;
  message: string;
}

const columns: ColumnDef<SyslogMessage>[] = [
  {
    accessorKey: "timestamp",
    header: "Timestamp",
    cell: ({ row }) => (
      <div className="font-mono text-xs">{row.getValue("timestamp")}</div>
    ),
  },
  {
    accessorKey: "severity",
    header: "Severity",
    cell: ({ row }) => {
      const severity = row.getValue("severity") as string;
      const colorMap: Record<string, string> = {
        emergency: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
        alert: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
        critical: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
        error: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
        warning: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
        notice: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
        info: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
        debug: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
      };
      return (
        <span
          className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ${
            colorMap[severity.toLowerCase()] || colorMap.info
          }`}
        >
          {severity}
        </span>
      );
    },
  },
  {
    accessorKey: "hostname",
    header: "Hostname",
  },
  {
    accessorKey: "facility",
    header: "Facility",
  },
  {
    accessorKey: "tag",
    header: "Tag",
  },
  {
    accessorKey: "message",
    header: "Message",
    cell: ({ row }) => (
      <div className="max-w-md truncate">{row.getValue("message")}</div>
    ),
  },
];

// Dummy data for demonstration
const dummyData: SyslogMessage[] = [
  {
    id: "1",
    timestamp: "2024-11-17T18:30:00Z",
    hostname: "server-01",
    facility: "daemon",
    severity: "info",
    tag: "sshd",
    message: "Accepted publickey for user from 192.168.1.100",
  },
  {
    id: "2",
    timestamp: "2024-11-17T18:30:05Z",
    hostname: "server-02",
    facility: "kern",
    severity: "warning",
    tag: "kernel",
    message: "Out of memory: Kill process 1234 (apache2)",
  },
  {
    id: "3",
    timestamp: "2024-11-17T18:30:10Z",
    hostname: "server-01",
    facility: "auth",
    severity: "error",
    tag: "sshd",
    message: "Failed password for invalid user admin from 192.168.1.200",
  },
];

export function SyslogTable() {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});

  const table = useReactTable({
    data: dummyData,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  });

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id} className="border-b bg-muted/50">
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className="px-4 py-3 text-left text-sm font-medium"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className="border-b transition-colors hover:bg-muted/50"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-4 py-3 text-sm">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    ))}
                  </tr>
                ))
              ) : (
                <tr>
                  <td
                    colSpan={columns.length}
                    className="h-24 text-center text-muted-foreground"
                  >
                    Aucun résultat.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          {table.getFilteredRowModel().rows.length} message(s) au total
        </div>
        <div className="flex items-center space-x-2">
          <button
            className="rounded border px-4 py-2 text-sm disabled:opacity-50"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Précédent
          </button>
          <button
            className="rounded border px-4 py-2 text-sm disabled:opacity-50"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Suivant
          </button>
        </div>
      </div>
    </div>
  );
}
