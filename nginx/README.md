# Nginx Configuration

This folder contains the nginx configuration for production deployment.

## Structure

```
nginx/
├── nginx.conf       # Main nginx configuration
├── ssl/             # SSL certificates (to be created)
│   ├── cert.pem
│   └── key.pem
└── README.md        # This file
```

## SSL Configuration

### Option 1: Self-signed Certificates (development/testing)

```bash
# Create ssl folder
mkdir -p nginx/ssl

# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=FR/ST=State/L=City/O=Organization/CN=syslog.example.com"
```

### Option 2: Let's Encrypt (recommended for production)

#### With Certbot Standalone

```bash
# Install certbot
sudo apt-get update
sudo apt-get install certbot

# Obtain a certificate
sudo certbot certonly --standalone \
  -d syslog.example.com \
  --email your@email.com \
  --agree-tos

# Copy certificates
sudo cp /etc/letsencrypt/live/syslog.example.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/syslog.example.com/privkey.pem nginx/ssl/key.pem
```

#### With Docker + Certbot

Add to `docker-compose.prod.yml`:

```yaml
services:
  certbot:
    image: certbot/certbot
    container_name: certbot
    volumes:
      - ./nginx/ssl:/etc/letsencrypt
      - ./nginx/www:/var/www/certbot
    command: certonly --webroot -w /var/www/certbot \
             --email your@email.com \
             -d syslog.example.com \
             --agree-tos
```

Then:

```bash
# Obtain certificate
docker-compose -f docker-compose.prod.yml run --rm certbot

# Restart nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

#### Automatic Renewal

Add a cron job:

```bash
# Edit crontab
crontab -e

# Add this line (renewal every Monday at 2 AM)
0 2 * * 1 docker-compose -f /path/to/docker-compose.prod.yml run --rm certbot renew && docker-compose -f /path/to/docker-compose.prod.yml restart nginx
```

### Option 3: Cloudflare SSL (if using Cloudflare)

1. Generate an Origin certificate in the Cloudflare dashboard
2. Download `cert.pem` and `key.pem`
3. Place in `nginx/ssl/`

## Configuration

### Change Domain Name

Edit `nginx/nginx.conf` and replace `syslog.example.com` with your domain:

```nginx
server_name syslog.example.com;  # Replace with your domain
```

### Test Configuration

```bash
# Test syntax
docker run --rm -v $(pwd)/nginx/nginx.conf:/etc/nginx/nginx.conf nginx nginx -t

# Or if nginx is installed locally
nginx -t -c nginx/nginx.conf
```

## Deployment

### Development (without SSL)

Edit `docker-compose.yml` to remove nginx service or use HTTP only.

### Production (with SSL)

```bash
# 1. Generate/obtain SSL certificates
# (see SSL Configuration section above)

# 2. Verify certificates exist
ls nginx/ssl/

# 3. Start with docker-compose production
docker-compose -f docker-compose.prod.yml up -d

# 4. Check logs
docker-compose -f docker-compose.prod.yml logs nginx
```

## Access

Once deployed with nginx:

- **Frontend**: https://syslog.example.com
- **API**: https://syslog.example.com/api
- **Health Check**: https://syslog.example.com/api/health

## Ports

- Port 80: HTTP (redirects to HTTPS)
- Port 443: HTTPS
- Port 514: Syslog (UDP/TCP) - exposed directly by backend

## Troubleshooting

### Error: "certificate verify failed"

SSL certificates are not valid or missing.

**Solution**: Generate certificates (see SSL Configuration section)

### Error: "502 Bad Gateway"

Backend or frontend is not responding.

**Solution**:
```bash
# Check all services are running
docker-compose -f docker-compose.prod.yml ps

# Check backend logs
docker-compose -f docker-compose.prod.yml logs backend

# Check frontend logs
docker-compose -f docker-compose.prod.yml logs frontend
```

### Error: "permission denied" for certificates

**Solution**:
```bash
# Set correct permissions
chmod 644 nginx/ssl/cert.pem
chmod 600 nginx/ssl/key.pem
```

### HTTPS redirect not working

**Solution**: Check that port 80 is properly mapped in docker-compose.prod.yml

## Optimizations

### Cache

Nginx cache is configured for static files:
- Files `/_next/static`: 60 minutes
- Headers `Cache-Control` with `immutable`

### Compression

Gzip is enabled for:
- text/plain, text/css, text/xml
- application/json, application/javascript
- Fonts and SVG

### Security

Security headers enabled:
- `Strict-Transport-Security` (HSTS)
- `X-Frame-Options`
- `X-Content-Type-Options`
- `X-XSS-Protection`

## Monitoring

### Nginx Logs

```bash
# View logs in real-time
docker-compose -f docker-compose.prod.yml logs -f nginx

# Access logs only
docker exec syslog-nginx tail -f /var/log/nginx/access.log

# Error logs only
docker exec syslog-nginx tail -f /var/log/nginx/error.log
```

### Nginx Status

```bash
# Check if nginx is running
docker-compose -f docker-compose.prod.yml ps nginx

# Reload configuration (without interruption)
docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload

# Test configuration
docker-compose -f docker-compose.prod.yml exec nginx nginx -t
```

## Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt](https://letsencrypt.org/)
- [Certbot](https://certbot.eff.org/)
- [SSL Labs Test](https://www.ssllabs.com/ssltest/)
