# Multi-stage Dockerfile for Syslog Visualizer Backend

# Stage 1: Build Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o syslog-server ./cmd/server

# Stage 2: Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/syslog-server .

# Create directory for database
RUN mkdir -p /data

# Expose ports
# 514: Syslog collector (UDP/TCP)
# 8080: API server
EXPOSE 514/udp 514/tcp 8080

# Environment variables with defaults
ENV RETENTION_PERIOD=7d
ENV CLEANUP_INTERVAL=1h
ENV ENABLE_RETENTION=true
ENV ENABLE_AUTH=false
ENV AUTH_USERS=""
ENV DB_PATH=/data/syslog.db

# Run the application
CMD ["./syslog-server"]
