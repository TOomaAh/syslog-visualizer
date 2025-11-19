"use client"

import * as React from "react"
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table"

import { DataTableToolbar } from "./data-table-toolbar"
import { getSeverityConfig, severityNames, SyslogMessage } from "./columns"

interface FilterOptions {
  hostnames: string[]
  tags: string[]
  facilities: number[]
  severities: number[]
}

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  columnFilters?: ColumnFiltersState
  onColumnFiltersChange?: (filters: ColumnFiltersState) => void
  filterOptions?: FilterOptions | null
  loadMoreRef?: React.RefObject<HTMLDivElement>
  loadingMore?: boolean
  hasMore?: boolean
}

export function DataTable<TData, TValue>({
  columns,
  data,
  columnFilters: externalColumnFilters,
  onColumnFiltersChange: externalOnColumnFiltersChange,
  filterOptions,
  loadMoreRef,
  loadingMore,
  hasMore,
}: DataTableProps<TData, TValue>) {
  const [rowSelection, setRowSelection] = React.useState({})
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({})
  const [internalColumnFilters, setInternalColumnFilters] =
    React.useState<ColumnFiltersState>([])
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnSizing, setColumnSizing] = React.useState({})
  const scrollContainerRef = React.useRef<HTMLDivElement>(null)
  const previousDataLengthRef = React.useRef(data.length)
  const scrollPositionRef = React.useRef(0)

  // Use external filters if provided, otherwise use internal state
  const columnFilters = externalColumnFilters ?? internalColumnFilters

  const setColumnFilters = React.useCallback(
    (updater: ColumnFiltersState | ((old: ColumnFiltersState) => ColumnFiltersState)) => {
      if (externalOnColumnFiltersChange) {
        const newValue = typeof updater === "function" ? updater(columnFilters) : updater
        externalOnColumnFiltersChange(newValue)
      } else {
        setInternalColumnFilters(updater)
      }
    },
    [externalOnColumnFiltersChange, columnFilters]
  )

  // Track scroll position continuously
  React.useEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return

    const handleScroll = () => {
      scrollPositionRef.current = container.scrollTop
    }

    container.addEventListener('scroll', handleScroll, { passive: true })
    return () => container.removeEventListener('scroll', handleScroll)
  }, [])

  // Restore scroll position after data refresh (only if data length is similar - meaning it's a refresh, not initial load or filter change)
  React.useLayoutEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return

    const dataDiff = Math.abs(data.length - previousDataLengthRef.current)

    // If data length changed by less than 50 rows, it's likely a refresh, restore scroll position
    if (dataDiff <= 50 && dataDiff > 0 && scrollPositionRef.current > 0) {
      container.scrollTop = scrollPositionRef.current
    }

    previousDataLengthRef.current = data.length
  }, [data])

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      columnSizing,
    },
    enableRowSelection: true,
    enableColumnResizing: true,
    columnResizeMode: "onChange",
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onColumnSizingChange: setColumnSizing,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    // Keep faceted models for filter UI
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    // Disable client-side filtering since we're doing server-side
    manualFiltering: true,
  })

  return (
    <div className="flex flex-col h-full min-h-0">
      <DataTableToolbar table={table} filterOptions={filterOptions} />
      <div className="rounded-xl border shadow-md bg-card flex flex-col flex-grow min-h-0 mt-4">
        <div ref={scrollContainerRef} className="overflow-auto flex-grow">
          <table className="w-full">
            <thead className="sticky top-0 z-10 bg-gray-100 border-b">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      colSpan={header.colSpan}
                      style={{
                        minWidth: `${header.column.columnDef.minSize}px`,
                        maxWidth: header.column.columnDef.maxSize ? `${header.column.columnDef.maxSize}px` : undefined,
                        width: header.column.columnDef.maxSize ? undefined : 'auto',
                        position: "relative",
                      }}
                      className="h-10 px-3 text-left align-middle font-semibold text-xs text-gray-600 [&:has([role=checkbox])]:pr-0"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                      {header.column.getCanResize() && (
                        <div
                          onMouseDown={header.getResizeHandler()}
                          onTouchStart={header.getResizeHandler()}
                          className={`absolute right-0 top-0 h-full w-2 cursor-col-resize select-none touch-none ${
                            header.column.getIsResizing()
                              ? "bg-primary opacity-100"
                              : "bg-border opacity-0 hover:opacity-100"
                          }`}
                          style={{ zIndex: 1 }}
                        />
                      )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows?.length ? (
                <>
                  {table.getRowModel().rows.map((row, index) => {
                    const message = row.original as SyslogMessage
                    const severity = message.severity
                    const severityName = severityNames[severity] || "unknown"
                    const severityConfig = getSeverityConfig(severityName, severity)

                    return (
                      <tr
                        key={row.id}
                        data-state={row.getIsSelected() && "selected"}
                        className={`
                          border-b border-l-4
                          transition-colors
                          ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}
                          hover:bg-gray-100
                          data-[state=selected]:bg-blue-50
                        `}
                        style={{
                          borderLeftColor: severityConfig.borderColor
                        }}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <td
                            key={cell.id}
                            style={{
                              minWidth: `${cell.column.columnDef.minSize}px`,
                              maxWidth: cell.column.columnDef.maxSize ? `${cell.column.columnDef.maxSize}px` : undefined,
                              width: cell.column.columnDef.maxSize ? undefined : 'auto',
                            }}
                            className="px-3 py-2 align-middle text-xs [&:has([role=checkbox])]:pr-0"
                          >
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </td>
                        ))}
                      </tr>
                    )
                  })}
                  {/* Infinite scroll trigger inside table */}
                  {loadMoreRef && (
                    <tr>
                      <td colSpan={columns.length}>
                        <div ref={loadMoreRef} className="h-20 flex items-center justify-center">
                          {loadingMore && (
                            <div className="text-sm text-muted-foreground">Loading more messages...</div>
                          )}
                          {!hasMore && data.length > 0 && (
                            <div className="text-sm text-muted-foreground">No more messages</div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ) : (
                <tr>
                  <td
                    colSpan={columns.length}
                    className="h-24 text-center text-gray-500"
                  >
                    <span className="text-sm">No syslog messages found</span>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="flex items-center justify-between px-3 py-2 border-t bg-gray-50 flex-shrink-0">
          <div className="flex-1 text-xs text-gray-600">
            <span>
              {table.getRowModel().rows.length} message{table.getRowModel().rows.length !== 1 ? 's' : ''}
            </span>
            {table.getSelectedRowModel().rows.length > 0 && (
              <span className="ml-2 text-gray-500">
                ({table.getSelectedRowModel().rows.length} selected)
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
