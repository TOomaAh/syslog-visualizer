# Authentication Examples

## 1. Starting with Authentication

### Without Authentication (default)
```bash
go run cmd/server/main.go
```
The API is publicly accessible.

### With Authentication
```bash
go run cmd/server/main.go -enable-auth -auth-users="admin:SecurePass123"
```

Expected output:
```
Syslog Visualizer starting...
Data retention enabled: keeping logs for 168h0m0s, cleanup every 1h0m0s
User created: admin (API Token: 8f7a3b2c1d5e4f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6)
Authentication enabled
Database initialized: syslog.db
Syslog Visualizer is running
  - Collector listening on :514 (UDP)
  - API server listening on :8080
Press Ctrl+C to stop
```

## 2. Testing with curl

### Without Authentication
```bash
# Health check (always accessible)
curl http://localhost:8080/api/health

# Fetch syslogs (accessible without auth if disabled)
curl http://localhost:8080/api/syslogs
```

### With Authentication - Session Cookie

```bash
# 1. Login and save session cookie
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"SecurePass123"}' \
  -c cookies.txt

# Response:
# {
#   "status": "success",
#   "username": "admin",
#   "apiToken": "8f7a3b2c1d5e4f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6",
#   "message": "Login successful"
# }

# 2. Access data with cookie
curl http://localhost:8080/api/syslogs -b cookies.txt

# 3. Logout
curl -X POST http://localhost:8080/api/auth/logout -b cookies.txt
```

### With Authentication - API Token (Bearer)

```bash
# Use the API token provided during login
export API_TOKEN="8f7a3b2c1d5e4f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"

curl http://localhost:8080/api/syslogs \
  -H "Authorization: Bearer $API_TOKEN"
```

### With Authentication - Basic Auth

```bash
# Use username:password directly
curl -u admin:SecurePass123 http://localhost:8080/api/syslogs
```

## 3. Python Script with Authentication

```python
#!/usr/bin/env python3
import requests

# Configuration
API_URL = "http://localhost:8080"
USERNAME = "admin"
PASSWORD = "SecurePass123"

# Method 1: Session with cookies
def with_session():
    session = requests.Session()

    # Login
    response = session.post(
        f"{API_URL}/api/auth/login",
        json={"username": USERNAME, "password": PASSWORD}
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Logged in as {data['username']}")
        print(f"API Token: {data['apiToken']}")

        # Fetch syslogs
        syslogs = session.get(f"{API_URL}/api/syslogs")
        print(f"Retrieved {len(syslogs.json())} syslogs")

        # Logout
        session.post(f"{API_URL}/api/auth/logout")
        print("Logged out")
    else:
        print(f"Login failed: {response.status_code}")

# Method 2: API Token
def with_api_token():
    # Login to get token
    response = requests.post(
        f"{API_URL}/api/auth/login",
        json={"username": USERNAME, "password": PASSWORD}
    )

    if response.status_code == 200:
        api_token = response.json()['apiToken']

        # Use token for subsequent requests
        headers = {"Authorization": f"Bearer {api_token}"}
        syslogs = requests.get(f"{API_URL}/api/syslogs", headers=headers)
        print(f"Retrieved {len(syslogs.json())} syslogs with API token")
    else:
        print(f"Login failed: {response.status_code}")

# Method 3: Basic Auth
def with_basic_auth():
    from requests.auth import HTTPBasicAuth

    response = requests.get(
        f"{API_URL}/api/syslogs",
        auth=HTTPBasicAuth(USERNAME, PASSWORD)
    )

    if response.status_code == 200:
        print(f"Retrieved {len(response.json())} syslogs with Basic Auth")
    else:
        print(f"Request failed: {response.status_code}")

if __name__ == "__main__":
    print("=== Session with cookies ===")
    with_session()

    print("\n=== API Token ===")
    with_api_token()

    print("\n=== Basic Auth ===")
    with_basic_auth()
```

## 4. Browser Usage

### Access without Authentication
1. Open http://localhost:3000
2. Direct access to dashboard

### Access with Authentication
1. Open http://localhost:3000
2. Automatic redirect to http://localhost:3000/login
3. Enter username and password
4. Redirect to dashboard
5. Session valid for 24 hours
6. "Logout" button in header

## 5. Environment Variables (Production)

```bash
# .env or export in shell

# Enable authentication
export ENABLE_AUTH=true

# Define users (never commit in plain text!)
export AUTH_USERS="admin:VerySecurePassword123,viewer:AnotherSecurePass456"

# Data retention
export RETENTION_PERIOD=30d
export CLEANUP_INTERVAL=24h

# Start server
go run cmd/server/main.go
```

## 6. Security

### Best Practices

DO:
- Use strong passwords (minimum 12 characters)
- Store credentials in environment variables
- Use HTTPS in production (reverse proxy nginx/caddy)
- Change default passwords
- Limit network access to server (firewall)

DO NOT:
- Never commit passwords in git
- Do not use simple passwords (admin/admin, password, etc.)
- Do not expose server directly on internet without HTTPS
- Do not share API tokens publicly

### HTTPS Configuration with Caddy (recommended)

```caddy
# Caddyfile
syslog.example.com {
    reverse_proxy localhost:8080

    # Caddy automatically handles Let's Encrypt certificates
}
```

```bash
# Start Caddy
caddy run --config Caddyfile
```

### HTTPS Configuration with nginx

```nginx
server {
    listen 443 ssl http2;
    server_name syslog.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 7. Multiple User Management

```bash
# Create multiple users with different roles
go run cmd/server/main.go \
  -enable-auth \
  -auth-users="admin:AdminPass123,viewer:ViewPass456,operator:OperPass789"
```

Each user receives:
- A password hashed with bcrypt
- A unique API token for scripts/tools
- Access to same endpoints (no permission system for now)

## 8. Troubleshooting

### Error: "Unauthorized"
```bash
# Check that authentication is enabled
curl http://localhost:8080/api/syslogs
# Response: 401 Unauthorized

# Solution: Login first
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Error: "Invalid credentials"
- Check username and password
- Credentials are case-sensitive
- Check for no spaces in AUTH_USERS

### Session Expired
- Sessions last 24 hours
- Re-login via `/api/auth/login`
- Use API token for long-running scripts
