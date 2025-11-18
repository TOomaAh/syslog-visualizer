# Syslog Visualizer - Frontend

Modern Next.js web interface for visualizing syslog messages with an interactive data table inspired by [data-table-filters](https://github.com/openstatusHQ/data-table-filters).

## Features

- **Instant search** in messages
- **Faceted filters** for Severity and Facility
- **Advanced pagination** with rows per page selection
- **Color-coded badges** by severity level
- **Relative timestamps** (e.g., "2 minutes ago")
- **Auto-refresh** every 5 seconds
- **Responsive interface** adapted to all screens
- **Row selection** with checkboxes

## Technology Stack

- **Next.js 14** - React framework with App Router
- **TypeScript** - Static typing
- **TanStack Table** - Powerful data table with sorting, filtering, faceted pagination
- **Tailwind CSS** - Utility styling
- **shadcn/ui** - Accessible and customizable UI components
- **Radix UI** - Headless UI primitives
- **date-fns** - Date manipulation

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

**Note:** The Go backend must be running on port 8080 for the API to work.

## Structure

```
web/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page with DataTable
│   └── globals.css        # Global styles + theme variables
├── components/
│   ├── ui/                # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── badge.tsx
│   │   ├── checkbox.tsx
│   │   ├── dropdown-menu.tsx
│   │   ├── popover.tsx
│   │   └── separator.tsx
│   └── syslog-table/      # Advanced table components
│       ├── data-table.tsx                # Main table
│       ├── data-table-toolbar.tsx        # Search bar + filters
│       ├── data-table-faceted-filter.tsx # Faceted filter (multi-select)
│       └── columns.tsx                   # Column definitions
├── lib/
│   └── utils.ts           # Utilities (cn function)
└── next.config.mjs        # Next.js config + API proxy
```

## Components

### DataTable

Main component that handles:
- State management (sorting, filters, pagination, selection)
- Table rendering with TanStack Table
- Pagination with complete controls
- Results count display

### DataTableToolbar

Toolbar with:
- **Search**: Input to search in messages
- **Severity Filters**: Multi-select with color-coded badges
- **Facility Filters**: Multi-select with badges
- **Reset Button**: Reset all filters

### DataTableFacetedFilter

Multi-select filter with:
- Radix UI Popover
- Checkboxes for multiple selection
- Results counter per option
- Badges for selected values
- "Clear filters" button

### Columns

Column definitions with:
- **Time**: Relative timestamp + full date
- **Severity**: Color-coded badge by level
- **Facility**: Facility name
- **Hostname**: Host name
- **Tag**: Tag + optional PID
- **Message**: Message truncated if too long

## Backend Integration

The frontend communicates with the Go backend via REST API:

```typescript
// Automatic fetch from /api/syslogs
const response = await fetch("/api/syslogs")
const data = await response.json()
```

The Next.js proxy (`next.config.mjs`) redirects:
- `/api/*` → `http://localhost:8080/api/*`

## Development

**Run complete server:**

```bash
# Terminal 1: Backend (collector + API)
cd ..
go run cmd/server/main.go

# Terminal 2: Frontend
npm run dev
```

**Test sending syslogs:**

```bash
# UDP
echo "<34>Oct 11 22:14:15 test su: test message" | nc -u localhost 514

# Check in frontend: http://localhost:3000
```

## Production Build

```bash
# Build
npm run build

# Run in production
npm run start
```

## Customization

Severity colors are defined in `columns.tsx`:

```typescript
const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    emergency: "bg-red-600 text-white",
    alert: "bg-orange-600 text-white",
    critical: "bg-orange-500 text-white",
    error: "bg-red-500 text-white",
    warning: "bg-yellow-500 text-black",
    notice: "bg-blue-500 text-white",
    info: "bg-green-500 text-white",
    debug: "bg-gray-500 text-white",
  }
  return colors[severity]
}
```

## Auto-refresh

The page automatically refreshes every 5 seconds:

```typescript
// In app/page.tsx
const interval = setInterval(fetchData, 5000)
```

You can adjust this value according to your needs.
