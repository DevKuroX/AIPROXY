# Deployment Guide

This guide covers production deployment of AI Proxy.

## Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- PostgreSQL 14+
- 2GB+ RAM minimum
- Ports 80/443 for reverse proxy, 20128 for direct access

## Quick Deployment

### Option 1: Binary Deployment

```bash
# Download and extract
wget https://github.com/DevKuroX/AIPROXY/releases/latest/download/ai_proxy_linux_amd64.tar.gz
tar xzf ai_proxy_linux_amd64.tar.gz

# Create environment file
cat > .env << 'EOF'
DATABASE_URL=postgres://user:pass@localhost:5432/ai_proxy?sslmode=disable
JWT_SECRET=$(openssl rand -hex 32)
ADMIN_PASSWORD=$(openssl rand -hex 16)
PORT=20128
EOF

# Run
./ai_proxy
```

### Option 2: Docker Deployment

```bash
# Clone and build
git clone https://github.com/DevKuroX/AIPROXY.git
cd AIPROXY

# Build image
docker build -t ai-proxy .

# Run with Docker
docker run -d \
  --name ai-proxy \
  -p 20128:20128 \
  -e DATABASE_URL="postgres://user:pass@host:5432/ai_proxy?sslmode=disable" \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  -e ADMIN_PASSWORD="$(openssl rand -hex 16)" \
  ai-proxy
```

### Option 3: Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ai_proxy
      POSTGRES_USER: ai_proxy
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ai_proxy"]
      interval: 5s
      timeout: 5s
      retries: 5

  ai-proxy:
    image: ai-proxy:latest
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "20128:20128"
    environment:
      DATABASE_URL: postgres://ai_proxy:${DB_PASSWORD}@postgres:5432/ai_proxy?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      ADMIN_PASSWORD: ${ADMIN_PASSWORD}
      PORT: 20128
    restart: unless-stopped

volumes:
  postgres_data:
```

Create `.env` file:

```bash
DB_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 32)
ADMIN_PASSWORD=$(openssl rand -hex 16)
```

Run:

```bash
docker-compose up -d
```

## PostgreSQL Setup

### Create Database

```bash
# As postgres user
sudo -u postgres psql

# Create user and database
CREATE USER ai_proxy WITH PASSWORD 'your-secure-password';
CREATE DATABASE ai_proxy OWNER ai_proxy;
GRANT ALL PRIVILEGES ON DATABASE ai_proxy TO ai_proxy;
```

### Connection String Format

```
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=MODE
```

Options for `sslmode`:
- `disable` - No SSL (development only)
- `require` - SSL required, no certificate verification
- `verify-full` - SSL with full certificate verification (production)

### Production PostgreSQL

```bash
# Recommended settings in postgresql.conf
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 768MB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2621kB
min_wal_size = 1GB
max_wal_size = 4GB
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret for JWT signing (32+ bytes) |
| `ADMIN_PASSWORD` | Yes | - | Dashboard admin password |
| `PORT` | No | `20128` | Server listen port |

Generate secure secrets:

```bash
# JWT secret (32 bytes hex)
openssl rand -hex 32

# Admin password (16 bytes hex)
openssl rand -hex 16
```

## Reverse Proxy Configuration

### Nginx

```nginx
upstream ai_proxy {
    server 127.0.0.1:20128;
    keepalive 32;
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # API endpoints
    location /v1/ {
        proxy_pass http://ai_proxy;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        # SSE streaming
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400s;
    }

    # Dashboard
    location / {
        proxy_pass http://ai_proxy;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddyfile
your-domain.com {
    reverse_proxy localhost:20128 {
        # SSE streaming support
        flush_interval -1
    }
}
```

Caddy automatically handles HTTPS with Let's Encrypt.

### Traefik

```yaml
# docker-compose.yml with Traefik
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.myresolver.acme.tlschallenge=true"
      - "--certificatesresolvers.myresolver.acme.email=your-email@example.com"
      - "--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json"
    ports:
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "./letsencrypt:/letsencrypt"

  ai-proxy:
    image: ai-proxy:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ai-proxy.rule=Host(`your-domain.com`)"
      - "traefik.http.routers.ai-proxy.tls.certresolver=myresolver"
      - "traefik.http.routers.ai-proxy.entrypoints=websecure"
    environment:
      DATABASE_URL: postgres://...
      JWT_SECRET: ...
      ADMIN_PASSWORD: ...
```

## Systemd Service

Create `/etc/systemd/system/ai-proxy.service`:

```ini
[Unit]
Description=AI Proxy
After=network.target postgresql.service

[Service]
Type=simple
User=ai-proxy
Group=ai-proxy
WorkingDirectory=/opt/ai-proxy
ExecStart=/opt/ai-proxy/ai_proxy
Restart=on-failure
RestartSec=5

# Environment
EnvironmentFile=/opt/ai-proxy/.env

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/opt/ai-proxy

# Resource limits
LimitNOFILE=65535
MemoryMax=1G

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable ai-proxy
sudo systemctl start ai-proxy
sudo systemctl status ai-proxy
```

## Production Checklist

- [ ] Strong JWT_SECRET (32+ random bytes)
- [ ] Strong ADMIN_PASSWORD
- [ ] PostgreSQL with SSL enabled
- [ ] Reverse proxy with HTTPS
- [ ] Rate limiting configured
- [ ] Firewall allowing only 80/443
- [ ] Log rotation configured
- [ ] Backup strategy for database
- [ ] Monitoring and alerts configured

## Monitoring

### Health Check

```bash
curl http://localhost:20128/health
# Returns: OK
```

### Prometheus Metrics

Metrics are exposed at `/metrics` (if enabled):

```bash
curl http://localhost:20128/metrics
```

Key metrics:
- `ai_proxy_requests_total` - Total requests by model, provider, status
- `ai_proxy_tokens_total` - Tokens used (prompt/completion)
- `ai_proxy_latency_seconds` - Request latency histogram

### Logging

Logs are written to stdout in JSON format. Use your preferred log aggregator:

```bash
# View logs
journalctl -u ai-proxy -f

# JSON format for structured logging
{"level":"info","time":"2024-01-15T10:30:00Z","msg":"request completed","method":"POST","path":"/v1/chat/completions","status":200,"duration_ms":1234}
```

## Backup

### Database Backup

```bash
# Full backup
pg_dump ai_proxy > backup_$(date +%Y%m%d).sql

# Restore
psql ai_proxy < backup_20240115.sql
```

### Automated Backup Script

```bash
#!/bin/bash
# /opt/ai-proxy/backup.sh
BACKUP_DIR="/var/backups/ai-proxy"
mkdir -p $BACKUP_DIR

pg_dump ai_proxy | gzip > $BACKUP_DIR/ai_proxy_$(date +%Y%m%d_%H%M%S).sql.gz

# Keep last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

Add to crontab:

```bash
0 2 * * * /opt/ai-proxy/backup.sh
```

## Troubleshooting

### Connection Refused

```bash
# Check if service is running
sudo systemctl status ai-proxy

# Check port
netstat -tlnp | grep 20128

# Check logs
journalctl -u ai-proxy -n 50
```

### Database Connection Issues

```bash
# Test connection
psql "$DATABASE_URL" -c "SELECT 1"

# Check PostgreSQL is running
sudo systemctl status postgresql
```

### High Memory Usage

```bash
# Check memory
free -m

# Check process memory
ps aux --sort=-%mem | head

# Restart service
sudo systemctl restart ai-proxy
```

## Scaling

### Horizontal Scaling

Run multiple instances behind a load balancer:

```
                    ┌─────────────┐
                    │ Load        │
                    │ Balancer    │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
    │Instance │       │Instance │       │Instance │
    │    1    │       │    2    │       │    3    │
    └─────────┘       └─────────┘       └─────────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    │  (Primary)  │
                    └─────────────┘
```

Key considerations:
- Use shared PostgreSQL instance
- Sticky sessions not required (stateless API)
- Configure health checks on load balancer
- Use connection pooling (pgx pool)
