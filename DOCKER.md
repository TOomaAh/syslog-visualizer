# Docker Guide - Syslog Visualizer

Complete guide for deploying Syslog Visualizer with Docker and Docker Compose.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Docker Architecture](#docker-architecture)
3. [Configuration](#configuration)
4. [Docker Commands](#docker-commands)
5. [Production](#production)
6. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+

### Quick Start

```bash
# 1. Clone the repository
git clone <repo-url>
cd syslog-visualizer

# 2. Create .env file (optional)
cp .env.example .env

# 3. Start all services
docker-compose up -d

# 4. Check logs
docker-compose logs -f

# 5. Access the application
# - Frontend: http://localhost:3000
# - API: http://localhost:8080
# - Syslog collector: UDP/TCP port 514
```

## Docker Architecture

The application uses a multi-container architecture:

```
┌─────────────────────────────────────────┐
│          Docker Compose                  │
├─────────────────┬───────────────────────┤
│   Backend       │      Frontend         │
│   (Go)          │      (Next.js)        │
│                 │                       │
│   Ports:        │      Ports:           │
│   - 514/udp     │      - 3000          │
│   - 514/tcp     │                       │
│   - 8080        │                       │
│                 │                       │
│   Volume:       │                       │
│   syslog-data   │                       │
└─────────────────┴───────────────────────┘
```

### Services

**Backend (syslog-backend)**
- Image: Alpine Linux + Go binary
- Collects syslogs (UDP/TCP port 514)
- REST API (port 8080)
- SQLite storage (persistent volume)
- Health check on `/api/health`

**Frontend (syslog-frontend)**
- Image: Alpine Linux + Node.js
- Next.js web interface (port 3000)
- Communicates with backend via Docker network
- Health check on home page

## Configuration

### Environment Variables

Create a `.env` file at the project root:

```bash
# Data retention
RETENTION_PERIOD=7d
CLEANUP_INTERVAL=1h
ENABLE_RETENTION=true

# Authentication (optional)
ENABLE_AUTH=false
AUTH_USERS=
```

### Configuration with Authentication

```bash
# .env
ENABLE_AUTH=true
AUTH_USERS=admin:SecurePassword123,viewer:ViewPass456
```

Then start:

```bash
docker-compose up -d
```

On startup, API tokens will be displayed in logs:

```bash
docker-compose logs backend | grep "API Token"
```

## Docker Commands

### Service Management

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Restart a specific service
docker-compose restart backend
docker-compose restart frontend

# View logs
docker-compose logs -f
docker-compose logs -f backend
docker-compose logs -f frontend

# View status
docker-compose ps
```

### Build and Rebuild

```bash
# Build all images
docker-compose build

# Build a specific image
docker-compose build backend
docker-compose build frontend

# Build without cache
docker-compose build --no-cache

# Rebuild and restart
docker-compose up -d --build
```

### Volume Management

```bash
# List volumes
docker volume ls

# Inspect data volume
docker volume inspect syslog-visualizer_syslog-data

# Backup volume
docker run --rm -v syslog-visualizer_syslog-data:/data -v $(pwd):/backup alpine tar czf /backup/syslog-backup.tar.gz /data

# Restore volume
docker run --rm -v syslog-visualizer_syslog-data:/data -v $(pwd):/backup alpine tar xzf /backup/syslog-backup.tar.gz -C /

# Remove volumes (WARNING: data loss)
docker-compose down -v
```

### Container Access

```bash
# Shell in backend
docker-compose exec backend sh

# Shell in frontend
docker-compose exec frontend sh

# View processes
docker-compose top

# Real-time stats
docker stats
```

## Production

### Optimal Configuration

**docker-compose.prod.yml**:

```yaml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: syslog-backend
    ports:
      - "514:514/udp"
      - "514:514/tcp"
      - "8080:8080"
    volumes:
      - syslog-data:/data
    environment:
      - RETENTION_PERIOD=30d
      - CLEANUP_INTERVAL=24h
      - ENABLE_RETENTION=true
      - ENABLE_AUTH=true
      - AUTH_USERS=${AUTH_USERS}
    restart: always
    networks:
      - syslog-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: syslog-frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080
      - NODE_ENV=production
    depends_on:
      - backend
    restart: always
    networks:
      - syslog-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # Reverse proxy (optional)
  nginx:
    image: nginx:alpine
    container_name: syslog-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - backend
      - frontend
    restart: always
    networks:
      - syslog-network

volumes:
  syslog-data:
    driver: local

networks:
  syslog-network:
    driver: bridge
```

Start with production file:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Reverse Proxy with Nginx

**nginx.conf**:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    upstream frontend {
        server frontend:3000;
    }

    server {
        listen 80;
        server_name syslog.example.com;

        # Redirect to HTTPS
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name syslog.example.com;

        ssl_certificate /etc/nginx/certs/cert.pem;
        ssl_certificate_key /etc/nginx/certs/key.pem;

        # Frontend
        location / {
            proxy_pass http://frontend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_cache_bypass $http_upgrade;
        }

        # API
        location /api {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### Monitoring and Logs

```bash
# Continuous logs with timestamps
docker-compose logs -f -t

# Filter by service
docker-compose logs -f backend
docker-compose logs -f frontend

# Follow last 100 lines
docker-compose logs --tail=100 -f

# Save logs
docker-compose logs > syslog-visualizer.log
```

## Building Individual Images

### Backend Only

```bash
# Build
docker build -t syslog-backend:latest .

# Run
docker run -d \
  --name syslog-backend \
  -p 514:514/udp \
  -p 514:514/tcp \
  -p 8080:8080 \
  -v syslog-data:/data \
  -e ENABLE_AUTH=true \
  -e AUTH_USERS="admin:password123" \
  syslog-backend:latest
```

### Frontend Only

```bash
# Build
cd web
docker build -t syslog-frontend:latest .

# Run
docker run -d \
  --name syslog-frontend \
  -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://backend:8080 \
  syslog-frontend:latest
```

## Troubleshooting

### Backend Won't Start

```bash
# Check logs
docker-compose logs backend

# Common errors:
# - Port 514 already in use: modify port mapping
# - Port 8080 already in use: modify port mapping
# - Permission issues: check volumes
```

### Frontend Cannot Connect to Backend

```bash
# Check that services are on the same network
docker network ls
docker network inspect syslog-visualizer_syslog-network

# Check environment variable
docker-compose exec frontend env | grep NEXT_PUBLIC_API_URL

# Test connectivity
docker-compose exec frontend wget -O- http://backend:8080/api/health
```

### Permission Issues

```bash
# Backend
docker-compose exec backend ls -la /data

# If issue, recreate volume
docker-compose down -v
docker-compose up -d
```

### Clean Up Docker

```bash
# Remove stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Clean everything (WARNING)
docker system prune -a --volumes
```

## Health Checks

Services include health checks:

```bash
# Check status
docker-compose ps

# Detailed status
docker inspect syslog-backend | grep -A 10 Health
docker inspect syslog-frontend | grep -A 10 Health

# Manual test
curl http://localhost:8080/api/health
curl http://localhost:3000
```

## Updating

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Verify everything works
docker-compose ps
docker-compose logs -f
```

## Security

### Best Practices

1. **Never commit .env file**
2. **Use strong passwords** in AUTH_USERS
3. **Use HTTPS** in production (nginx + Let's Encrypt)
4. **Limit port exposure** (firewall)
5. **Monitor logs** regularly
6. **Keep Docker up to date**
7. **Scan images** for vulnerabilities

```bash
# Scan an image
docker scan syslog-backend:latest
docker scan syslog-frontend:latest
```

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Next.js Docker](https://nextjs.org/docs/deployment#docker-image)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
