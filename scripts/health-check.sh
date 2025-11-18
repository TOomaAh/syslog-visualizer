#!/bin/bash

# Health Check Script for Syslog Visualizer
# Usage: ./scripts/health-check.sh

set -e

echo "=== Syslog Visualizer Health Check ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"

# Check if running in Docker
if [ -f /.dockerenv ]; then
    BACKEND_URL="http://backend:8080"
    FRONTEND_URL="http://frontend:3000"
fi

# Function to check URL
check_url() {
    local url=$1
    local name=$2

    if curl -sf "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}[OK]${NC} $name: OK"
        return 0
    else
        echo -e "${RED}[ERROR]${NC} $name: FAILED"
        return 1
    fi
}

# Function to check JSON response
check_json() {
    local url=$1
    local name=$2

    response=$(curl -sf "$url" 2>&1)
    if [ $? -eq 0 ]; then
        if echo "$response" | jq . > /dev/null 2>&1; then
            echo -e "${GREEN}[OK]${NC} $name: OK (valid JSON)"
            return 0
        else
            echo -e "${YELLOW}[WARNING]${NC} $name: OK (invalid JSON)"
            return 1
        fi
    else
        echo -e "${RED}[ERROR]${NC} $name: FAILED"
        return 1
    fi
}

# Check Docker services
echo "1. Checking Docker Services..."
if command -v docker-compose &> /dev/null; then
    if docker-compose ps | grep -q "Up"; then
        echo -e "${GREEN}[OK]${NC} Docker Compose services are running"
        docker-compose ps
    else
        echo -e "${YELLOW}[WARNING]${NC} Docker Compose services may not be running"
    fi
else
    echo -e "${YELLOW}[WARNING]${NC} docker-compose not found, skipping service check"
fi
echo ""

# Check Backend
echo "2. Checking Backend ($BACKEND_URL)..."
check_json "$BACKEND_URL/api/health" "Backend Health Endpoint"
echo ""

# Check Frontend
echo "3. Checking Frontend ($FRONTEND_URL)..."
check_url "$FRONTEND_URL" "Frontend Home Page"
echo ""

# Check Ports
echo "4. Checking Ports..."
if command -v netstat &> /dev/null; then
    if netstat -tuln | grep -q ":514"; then
        echo -e "${GREEN}[OK]${NC} Port 514 (Syslog): LISTENING"
    else
        echo -e "${RED}[ERROR]${NC} Port 514 (Syslog): NOT LISTENING"
    fi

    if netstat -tuln | grep -q ":8080"; then
        echo -e "${GREEN}[OK]${NC} Port 8080 (API): LISTENING"
    else
        echo -e "${RED}[ERROR]${NC} Port 8080 (API): NOT LISTENING"
    fi

    if netstat -tuln | grep -q ":3000"; then
        echo -e "${GREEN}[OK]${NC} Port 3000 (Frontend): LISTENING"
    else
        echo -e "${RED}[ERROR]${NC} Port 3000 (Frontend): NOT LISTENING"
    fi
else
    echo -e "${YELLOW}[WARNING]${NC} netstat not found, skipping port check"
fi
echo ""

# Check Database
echo "5. Checking Database..."
if [ -f "syslog.db" ]; then
    size=$(du -h syslog.db | cut -f1)
    echo -e "${GREEN}[OK]${NC} Database file exists (size: $size)"
else
    echo -e "${YELLOW}[WARNING]${NC} Database file not found (may be in Docker volume)"
fi
echo ""

# Test sending a syslog message
echo "6. Testing Syslog Collection..."
if command -v nc &> /dev/null; then
    test_message="<34>$(date '+%b %d %H:%M:%S') test-host health-check: Test message from health-check script"

    if echo "$test_message" | nc -u -w 1 localhost 514 2>/dev/null; then
        echo -e "${GREEN}[OK]${NC} Successfully sent test syslog message (UDP)"
        sleep 1

        # Check if message was received
        if curl -sf "$BACKEND_URL/api/syslogs" | grep -q "health-check" 2>/dev/null; then
            echo -e "${GREEN}[OK]${NC} Test message appears in API response"
        else
            echo -e "${YELLOW}[WARNING]${NC} Test message not found in API response (may require authentication)"
        fi
    else
        echo -e "${RED}[ERROR]${NC} Failed to send test syslog message"
    fi
else
    echo -e "${YELLOW}[WARNING]${NC} netcat (nc) not found, skipping syslog test"
fi
echo ""

# Summary
echo "=== Health Check Complete ==="
echo ""
echo "URLs to access:"
echo "  - Frontend: $FRONTEND_URL"
echo "  - API:      $BACKEND_URL/api/health"
echo "  - Syslogs:  $BACKEND_URL/api/syslogs"
echo ""
