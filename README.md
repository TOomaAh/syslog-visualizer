# Syslog Visualizer

A full-stack project for collecting and visualizing syslog messages in real-time.

## Features

- **Collector** (Go): Listens and collects syslog messages (UDP/TCP)
- **Parser** (Go): Parses syslog messages according to RFC 3164 and RFC 5424
- **Storage** (Go): Message storage with SQLite and GORM
- **Retention** (Go): Automatic deletion of old data (configurable)
- **Authentication** (Go): Authentication system with sessions and API tokens
- **API Backend** (Go): REST API server to expose logs
- **Visualizer** (Next.js): Modern web interface with interactive data table, sorting, filtering and pagination

## Technology Stack

**Backend:**
- Go 1.21+
- REST API with net/http
- SQLite storage with GORM
- Automatic cleanup of old data

**Frontend:**
- Next.js 14 (App Router)
- React + TypeScript
- TanStack Table (sorting, filters, pagination)
- Tailwind CSS + shadcn/ui
- Inspired by [data-table-filters](https://github.com/openstatusHQ/data-table-filters)

## Project Structure

```
.
├── cmd/
│   ├── server/          # Main server (collector + REST API)
│   ├── collector/       # (Deprecated) Standalone collector
│   └── visualizer/      # (Deprecated) Standalone API
├── internal/
│   ├── collector/       # UDP/TCP collection logic
│   ├── framing/         # TCP framing (RFC 6587)
│   ├── parser/          # RFC 3164/5424 parser
│   └── storage/         # Interface and storage backends
├── pkg/
│   └── syslog/          # Syslog constants and utilities
├── web/                 # Next.js application
│   ├── app/             # Pages and layouts (App Router)
│   ├── components/      # React components
│   └── public/          # Static assets
└── configs/             # YAML configuration
```

## Prerequisites

- Go 1.24 or higher
- Node.js 18+ and npm

## Installation

**Backend:**
```bash
go mod download
```

**Frontend:**
```bash
cd web
npm install
```

## Usage

### Docker (Recommended)

The simplest way to deploy the application:

```bash
# Quick start with Docker Compose
docker-compose up -d

# Access:
# - Frontend: http://localhost:3000
# - API: http://localhost:8080
# - Syslog: UDP/TCP port 514

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

For more details, see [DOCKER.md](./DOCKER.md).

### Development

**1. Start Backend Server** (terminal 1)
```bash
go run cmd/server/main.go
```
The server starts:
- Syslog collector on port 514 (UDP)
- REST API on http://localhost:8080

**2. Start Frontend** (terminal 2)
```bash
cd web
npm run dev
```
The web interface will be available at http://localhost:3000

### Production

**Build the server:**
```bash
go build -o bin/syslog-visualizer cmd/server/main.go
```

**Run the server:**
```bash
# Linux/macOS (requires sudo for port 514)
sudo ./bin/syslog-visualizer

# Or use a non-privileged port (>1024) by modifying the config
./bin/syslog-visualizer
```

**Run the frontend:**
```bash
cd web
npm run build
npm run start
```

## Tests

**Backend:**
```bash
# Run all Go tests
go test ./...

# Tests with coverage
go test -cover ./...
```

**Frontend:**
```bash
cd web
npm run lint
```

## Configuration

Copy the example configuration file:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Edit `configs/config.yaml` according to your needs.

### Data Retention Configuration

The server supports automatic cleanup of old data to prevent the database from growing indefinitely.

**Via command-line flags:**
```bash
# Keep logs for 7 days, cleanup every hour (default)
go run cmd/server/main.go

# Keep logs for 24 hours, cleanup every 30 minutes
go run cmd/server/main.go -retention=24h -cleanup-interval=30m

# Keep logs for 30 days, cleanup every 6 hours
go run cmd/server/main.go -retention=30d -cleanup-interval=6h

# Disable automatic retention (logs kept indefinitely)
go run cmd/server/main.go -enable-retention=false
```

**Via environment variables:**
```bash
# Linux/macOS
export RETENTION_PERIOD=14d
export CLEANUP_INTERVAL=2h
export ENABLE_RETENTION=true
go run cmd/server/main.go

# Windows
set RETENTION_PERIOD=14d
set CLEANUP_INTERVAL=2h
set ENABLE_RETENTION=true
go run cmd/server/main.go
```

**Available options:**
- `RETENTION_PERIOD` / `-retention`: Log retention duration
  - Format: `24h`, `7d`, `30d`, etc.
  - Default: `7d` (7 days)
- `CLEANUP_INTERVAL` / `-cleanup-interval`: Automatic cleanup frequency
  - Format: `30m`, `1h`, `6h`, etc.
  - Default: `1h` (1 hour)
- `ENABLE_RETENTION` / `-enable-retention`: Enable/disable retention
  - Values: `true` or `false`
  - Default: `true`

**Usage examples:**
```bash
# Production: keep 30 days, daily cleanup
go run cmd/server/main.go -retention=30d -cleanup-interval=24h

# Development: keep 1 day, frequent cleanup
go run cmd/server/main.go -retention=24h -cleanup-interval=10m

# Tests: disable retention
go run cmd/server/main.go -enable-retention=false
```

### Authentication Configuration

The server supports authentication to secure access to the API and web interface.

**Via command-line flags:**
```bash
# Start with authentication enabled
go run cmd/server/main.go -enable-auth -auth-users="admin:password123,user:pass456"

# With a single user
go run cmd/server/main.go -enable-auth -auth-users="admin:mySecurePassword"

# Without authentication (default - public access)
go run cmd/server/main.go
```

**Via environment variables:**
```bash
# Linux/macOS
export ENABLE_AUTH=true
export AUTH_USERS="admin:password123,viewer:viewpass"
go run cmd/server/main.go

# Windows
set ENABLE_AUTH=true
set AUTH_USERS="admin:password123,viewer:viewpass"
go run cmd/server/main.go
```

**Available options:**
- `ENABLE_AUTH` / `-enable-auth`: Enable/disable authentication
  - Values: `true` or `false`
  - Default: `false` (authentication disabled)
- `AUTH_USERS` / `-auth-users`: List of users in `username:password` format
  - Format: `"user1:pass1,user2:pass2"`
  - Comma-separated for multiple users
  - **Required** if authentication is enabled

**On startup with authentication, the server displays:**
```
User created: admin (API Token: 8f7a3b2c1d5e4f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6)
User created: viewer (API Token: 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7)
Authentication enabled
```

**Supported authentication methods:**

1. **Session Cookie** (for web)
   - Login via `/api/auth/login` with username/password
   - Session cookie valid for 24 hours
   - Used automatically by web interface

2. **API Token** (for scripts/tools)
   ```bash
   # Use Bearer token
   curl -H "Authorization: Bearer <API_TOKEN>" http://localhost:8080/api/syslogs
   ```

3. **Basic Auth** (for compatibility)
   ```bash
   curl -u admin:password123 http://localhost:8080/api/syslogs
   ```

**Example with curl:**
```bash
# 1. Login and get API token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' \
  -c cookies.txt

# Response:
# {
#   "status": "success",
#   "username": "admin",
#   "apiToken": "8f7a3b2c1d5e4f6g...",
#   "message": "Login successful"
# }

# 2. Access with session cookie
curl http://localhost:8080/api/syslogs -b cookies.txt

# 3. Access with API token
curl http://localhost:8080/api/syslogs \
  -H "Authorization: Bearer 8f7a3b2c1d5e4f6g..."

# 4. Logout
curl -X POST http://localhost:8080/api/auth/logout -b cookies.txt
```

**Web Interface:**
- If authentication is enabled, the interface automatically redirects to `/login`
- After login, the session is maintained for 24 hours
- "Logout" button available in the header

## API Endpoints

**Public endpoints:**
- `GET /api/health` - Server health check
- `POST /api/auth/login` - Login (returns session cookie and API token)
- `POST /api/auth/logout` - Logout (invalidates session)

**Protected endpoints** (requires authentication if enabled):
- `GET /api/syslogs` - Retrieve syslog messages (default limit: 100)

## Architecture

**Unified Backend Server** (`cmd/server/main.go`):
- Collects syslogs (UDP/TCP on port 514)
- Parses messages (RFC 3164/5424)
- Stores in SQLite with GORM (`syslog.db` file)
- Automatic cleanup of old data (configurable)
- Exposes a REST API (port 8080)

**Frontend** (Next.js on port 3000):
- Communicates with API via proxy (`/api/*` → `localhost:8080`)
- Interactive data table with TanStack Table
- Filters, sorting, pagination

**Data Flow:**
```
Syslog Source → Collector (UDP/TCP:514) → Parser → Storage ← API (:8080) ← Frontend (:3000)
```

## Documentation

- **[CLAUDE.md](./CLAUDE.md)** - Detailed development guide
- **[DOCKER.md](./DOCKER.md)** - Complete Docker and deployment guide
- **[examples/auth-example.md](./examples/auth-example.md)** - Authentication examples
- **[nginx/README.md](./nginx/README.md)** - Nginx configuration and SSL
