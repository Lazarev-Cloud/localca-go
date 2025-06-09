# LocalCA Deployment Guide

This guide covers all deployment options for LocalCA, from simple Docker setups to production deployments with enhanced storage and monitoring.

## Quick Start Deployment

### Docker Compose (Recommended)

The fastest way to get LocalCA running is with Docker Compose:

```bash
# Clone the repository
git clone https://github.com/Lazarev-Cloud/localca-go.git
cd localca-go

# Start with default configuration
docker-compose up -d

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
```

### Environment Configuration

Create a `.env` file for custom configuration:

```bash
# Copy example configuration
cp .env.example .env

# Edit configuration
nano .env
```

Basic `.env` configuration:
```bash
# Core Configuration
CA_NAME=MyLocalCA
CA_KEY_PASSWORD=secure-ca-password
ORGANIZATION=My Organization
COUNTRY=US

# Network Configuration
LISTEN_ADDR=:8080
NEXT_PUBLIC_API_URL=http://localhost:8080

# Security
TLS_ENABLED=false
SESSION_SECRET=your-secure-session-secret
```

## Production Deployment

### Enhanced Storage Configuration

For production deployments, enable enhanced storage features:

```bash
# Enhanced Storage Configuration
DATABASE_ENABLED=true
DATABASE_URL=postgres://localca:secure_password@postgres:5432/localca
DATABASE_MAX_CONNECTIONS=25
DATABASE_SSL_MODE=require

# S3/MinIO Object Storage
S3_ENABLED=true
S3_ENDPOINT=http://minio:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=secure_minio_password
S3_BUCKET=localca-certificates
S3_REGION=us-east-1
S3_SSL=false

# Redis/KeyDB Caching
CACHE_ENABLED=true
REDIS_URL=redis://keydb:6379
REDIS_PASSWORD=secure_redis_password
CACHE_TTL_DEFAULT=3600

# Email Notifications
EMAIL_NOTIFY=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourdomain.com

# Logging
LOG_FORMAT=json
LOG_LEVEL=info
AUDIT_ENABLED=true

# Security
TLS_ENABLED=true
TLS_CERT_FILE=/certs/server.crt
TLS_KEY_FILE=/certs/server.key
```

### Production Docker Compose

Use the production Docker Compose configuration:

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_ENABLED=true
      - DATABASE_URL=postgres://localca:${POSTGRES_PASSWORD}@postgres:5432/localca
      - S3_ENABLED=true
      - S3_ENDPOINT=http://minio:9000
      - CACHE_ENABLED=true
      - REDIS_URL=redis://keydb:6379
    volumes:
      - ./data:/app/data
      - ./certs:/certs:ro
    depends_on:
      - postgres
      - minio
      - keydb
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080
    depends_on:
      - backend
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=localca
      - POSTGRES_USER=localca
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U localca"]
      interval: 30s
      timeout: 10s
      retries: 3

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=${MINIO_ROOT_USER}
      - MINIO_ROOT_PASSWORD=${MINIO_ROOT_PASSWORD}
    volumes:
      - minio_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3

  keydb:
    image: eqalpha/keydb:latest
    command: keydb-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - keydb_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "keydb-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - frontend
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  minio_data:
  keydb_data:

networks:
  default:
    driver: bridge
```

### Nginx Configuration

Create `nginx.conf` for reverse proxy and SSL termination:

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

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server.key;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header Strict-Transport-Security "max-age=63072000" always;
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Frontend
        location / {
            proxy_pass http://frontend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Backend API
        location /api/ {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # ACME server
        location /.well-known/acme-challenge/ {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## Kubernetes Deployment

### Kubernetes Manifests

Deploy LocalCA on Kubernetes with the following manifests:

#### Namespace
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: localca
```

#### ConfigMap
```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: localca-config
  namespace: localca
data:
  CA_NAME: "LocalCA"
  ORGANIZATION: "LocalCA Organization"
  COUNTRY: "US"
  DATABASE_ENABLED: "true"
  S3_ENABLED: "true"
  CACHE_ENABLED: "true"
  EMAIL_NOTIFY: "true"
  LOG_FORMAT: "json"
  LOG_LEVEL: "info"
```

#### Secrets
```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: localca-secrets
  namespace: localca
type: Opaque
data:
  CA_KEY_PASSWORD: <base64-encoded-password>
  DATABASE_URL: <base64-encoded-database-url>
  S3_ACCESS_KEY: <base64-encoded-s3-access-key>
  S3_SECRET_KEY: <base64-encoded-s3-secret-key>
  REDIS_PASSWORD: <base64-encoded-redis-password>
  SMTP_PASSWORD: <base64-encoded-smtp-password>
```

#### Deployment
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: localca-backend
  namespace: localca
spec:
  replicas: 2
  selector:
    matchLabels:
      app: localca-backend
  template:
    metadata:
      labels:
        app: localca-backend
    spec:
      containers:
      - name: backend
        image: localca/backend:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: localca-config
        - secretRef:
            name: localca-secrets
        volumeMounts:
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: localca-data
```

#### Service
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: localca-backend
  namespace: localca
spec:
  selector:
    app: localca-backend
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

#### Ingress
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: localca-ingress
  namespace: localca
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - localca.yourdomain.com
    secretName: localca-tls
  rules:
  - host: localca.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: localca-frontend
            port:
              number: 3000
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: localca-backend
            port:
              number: 8080
```

## Standalone Deployment

### Binary Deployment

Build and deploy LocalCA as standalone binaries:

```bash
# Build backend
go build -o localca-go main.go

# Build frontend
npm install
npm run build

# Create systemd service
sudo tee /etc/systemd/system/localca.service > /dev/null <<EOF
[Unit]
Description=LocalCA Certificate Authority
After=network.target

[Service]
Type=simple
User=localca
WorkingDirectory=/opt/localca
ExecStart=/opt/localca/localca-go
Restart=always
RestartSec=5
Environment=DATA_DIR=/var/lib/localca
Environment=LISTEN_ADDR=:8080

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl enable localca
sudo systemctl start localca
```

### Frontend Deployment

Deploy the frontend separately:

```bash
# Build frontend for production
npm run build

# Serve with nginx
sudo tee /etc/nginx/sites-available/localca > /dev/null <<EOF
server {
    listen 80;
    server_name localca.yourdomain.com;

    root /var/www/localca;
    index index.html;

    location / {
        try_files \$uri \$uri/ @nextjs;
    }

    location @nextjs {
        proxy_pass http://localhost:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/localca /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Security Hardening

### SSL/TLS Configuration

Generate SSL certificates for production:

```bash
# Generate self-signed certificate for testing
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# Or use Let's Encrypt with certbot
sudo certbot certonly --nginx -d localca.yourdomain.com
```

### Firewall Configuration

Configure firewall rules:

```bash
# UFW configuration
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# iptables configuration
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -A INPUT -j DROP
```

### User and Permissions

Create dedicated user for LocalCA:

```bash
# Create localca user
sudo useradd -r -s /bin/false localca

# Create directories
sudo mkdir -p /opt/localca /var/lib/localca /var/log/localca

# Set permissions
sudo chown -R localca:localca /opt/localca /var/lib/localca /var/log/localca
sudo chmod 750 /opt/localca /var/lib/localca /var/log/localca
```

## Monitoring and Logging

### Prometheus Monitoring

Configure Prometheus monitoring:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'localca'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/metrics'
    scrape_interval: 30s
```

### Log Aggregation

Configure log aggregation with ELK stack:

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/localca/*.log
  fields:
    service: localca
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]

setup.kibana:
  host: "kibana:5601"
```

### Health Checks

Implement health check monitoring:

```bash
#!/bin/bash
# health-check.sh

# Check backend health
if ! curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "Backend health check failed"
    exit 1
fi

# Check frontend health
if ! curl -f http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "Frontend health check failed"
    exit 1
fi

# Check database connectivity
if ! pg_isready -h localhost -p 5432 -U localca > /dev/null 2>&1; then
    echo "Database health check failed"
    exit 1
fi

echo "All health checks passed"
```

## Backup and Recovery

### Database Backup

Automated database backup:

```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/var/backups/localca"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/localca_backup_$DATE.sql"

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
pg_dump -h localhost -U localca localca > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Remove old backups (keep 30 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Database backup completed: $BACKUP_FILE.gz"
```

### Certificate Backup

Backup certificate data:

```bash
#!/bin/bash
# backup-certificates.sh

BACKUP_DIR="/var/backups/localca"
DATE=$(date +%Y%m%d_%H%M%S)
DATA_DIR="/var/lib/localca"

# Create backup
tar -czf "$BACKUP_DIR/certificates_backup_$DATE.tar.gz" -C "$DATA_DIR" .

# Remove old backups
find $BACKUP_DIR -name "certificates_backup_*.tar.gz" -mtime +30 -delete

echo "Certificate backup completed: certificates_backup_$DATE.tar.gz"
```

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database connectivity
   pg_isready -h localhost -p 5432 -U localca
   
   # Check database logs
   docker-compose logs postgres
   ```

2. **Storage Issues**
   ```bash
   # Check MinIO connectivity
   curl -f http://localhost:9000/minio/health/live
   
   # Check storage permissions
   ls -la /var/lib/localca
   ```

3. **Cache Issues**
   ```bash
   # Check Redis connectivity
   redis-cli ping
   
   # Check cache statistics
   curl http://localhost:8080/api/statistics
   ```

### Log Analysis

Check application logs:

```bash
# Backend logs
docker-compose logs backend

# Frontend logs
docker-compose logs frontend

# System logs
journalctl -u localca -f
```

### Performance Tuning

Optimize performance:

```bash
# Database optimization
# Increase shared_buffers in postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB

# Redis optimization
# Increase maxmemory in redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
```

## Maintenance

### Regular Maintenance Tasks

1. **Update Dependencies**
   ```bash
   # Update Go dependencies
   go mod tidy
   go mod download
   
   # Update Node.js dependencies
   npm update
   ```

2. **Certificate Rotation**
   ```bash
   # Rotate CA certificate (if needed)
   # This should be done carefully and rarely
   ```

3. **Database Maintenance**
   ```bash
   # Vacuum database
   psql -h localhost -U localca -c "VACUUM ANALYZE;"
   
   # Update statistics
   psql -h localhost -U localca -c "ANALYZE;"
   ```

4. **Log Rotation**
   ```bash
   # Configure logrotate
   sudo tee /etc/logrotate.d/localca > /dev/null <<EOF
   /var/log/localca/*.log {
       daily
       rotate 30
       compress
       delaycompress
       missingok
       notifempty
       create 644 localca localca
   }
   EOF
   ```

This deployment guide provides comprehensive instructions for deploying LocalCA in various environments, from development to production. Choose the deployment method that best fits your infrastructure and security requirements.