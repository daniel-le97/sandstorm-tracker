# Docker Setup Guide for sandstorm-tracker

## Quick Start

### 1. Build and Run with Docker Compose

```bash
# Navigate to project root
cd /path/to/sandstorm-tracker

# Start the container
docker-compose up -d

# View logs
docker-compose logs -f sandstorm-tracker

# Stop the container
docker-compose down
```

### 2. First-Time Setup

```bash
# Create a config file from the example
cp sandstorm-tracker.example.yml sandstorm-tracker.yml

# Edit the config with your server paths
nano sandstorm-tracker.yml

# Start Docker Compose
docker-compose up -d

# Access the PocketBase Admin UI
# Open browser: http://localhost:8090/_/
```

### 3. Access the Application

- **Admin UI**: http://localhost:8090/\_/
- **API**: http://localhost:8090/api/
- **Data Directory**: `./pb_data/` (mounted volume)
- **Logs**: `./logs/` (mounted volume)

---

## Configuration with Docker

### Using Environment Variables

Edit `docker-compose.yml` to set environment variables:

```yaml
environment:
  - TZ=UTC
  - DEBUG=false
  # Custom logging level
  - LOG_LEVEL=info
```

### Using Config File

Mount your config file as a volume:

```yaml
volumes:
  - ./sandstorm-tracker.yml:/app/sandstorm-tracker.yml:ro
```

### Mounting Game Server Logs

To watch game server logs from host machine:

```yaml
volumes:
  # Assuming game logs are at /opt/sandstorm/Insurgency/Saved/Logs
  - /opt/sandstorm/Insurgency/Saved/Logs:/mnt/game_logs:ro
```

Then update your `sandstorm-tracker.yml`:

```yaml
servers:
  - name: "Main Server"
    logPath: "/mnt/game_logs" # Inside container path
    rconAddress: "192.168.1.100:27015"
    rconPassword: "your-password"
```

---

## Advanced Docker Setup

### 1. Multi-Container Orchestration (Kubernetes)

Create `k8s-deployment.yml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandstorm-tracker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sandstorm-tracker
  template:
    metadata:
      labels:
        app: sandstorm-tracker
    spec:
      containers:
        - name: sandstorm-tracker
          image: localhost:5000/sandstorm-tracker:latest
          ports:
            - containerPort: 8090
          volumeMounts:
            - name: pb-data
              mountPath: /app/pb_data
            - name: config
              mountPath: /app/sandstorm-tracker.yml
              subPath: sandstorm-tracker.yml
            - name: logs
              mountPath: /app/logs
          resources:
            requests:
              memory: "256Mi"
              cpu: "500m"
            limits:
              memory: "512Mi"
              cpu: "1000m"
          livenessProbe:
            httpGet:
              path: /api/health
              port: 8090
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /api/health
              port: 8090
            initialDelaySeconds: 5
            periodSeconds: 10
      volumes:
        - name: pb-data
          persistentVolumeClaim:
            claimName: sandstorm-pvc
        - name: config
          configMap:
            name: sandstorm-config
        - name: logs
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: sandstorm-tracker-service
spec:
  type: LoadBalancer
  ports:
    - port: 8090
      targetPort: 8090
  selector:
    app: sandstorm-tracker
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: sandstorm-config
data:
  sandstorm-tracker.yml: |
    servers:
      - name: "Main Server"
        logPath: "/var/log/sandstorm"
        rconAddress: "game-server:27015"
        rconPassword: "password"
```

Deploy:

```bash
kubectl apply -f k8s-deployment.yml
```

### 2. Docker Swarm Stack

Create `docker-stack.yml`:

```yaml
version: "3.8"

services:
  sandstorm-tracker:
    image: sandstorm-tracker:latest
    ports:
      - "8090:8090"
    volumes:
      - pb_data:/app/pb_data
      - ./logs:/app/logs
      - ./sandstorm-tracker.yml:/app/sandstorm-tracker.yml:ro
    environment:
      - TZ=UTC
    restart: unless-stopped
    deploy:
      replicas: 1
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s

volumes:
  pb_data:
    driver: local
```

Deploy:

```bash
docker stack deploy -c docker-stack.yml sandstorm
```

### 3. Local Registry (for CI/CD)

```bash
# Start local registry
docker run -d -p 5000:5000 --name registry registry:2

# Build and push image
docker build -t localhost:5000/sandstorm-tracker:latest .
docker push localhost:5000/sandstorm-tracker:latest

# Use in docker-compose
# image: localhost:5000/sandstorm-tracker:latest
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs sandstorm-tracker

# Verify config file exists
docker-compose exec sandstorm-tracker ls -la /app/sandstorm-tracker.yml

# Check permissions
docker-compose exec sandstorm-tracker ls -la /app/pb_data
```

### Permission Issues

The container runs as non-root user (UID 1000). If you have permission issues:

```bash
# Fix volume permissions on host
sudo chown -R 1000:1000 ./pb_data
sudo chown -R 1000:1000 ./logs
```

### Database Locked

```bash
# Remove and recreate database
docker-compose down -v  # Warning: deletes data!
docker-compose up -d
```

### Network Issues

```bash
# Test DNS resolution
docker-compose exec sandstorm-tracker nslookup game-server.local

# Test connectivity to RCON
docker-compose exec sandstorm-tracker telnet game-server 27015
```

### Memory Leaks

Monitor container memory:

```bash
# Check resource usage
docker stats sandstorm-tracker

# If memory grows unbounded, add restart on limit
# Edit docker-compose.yml and set memory limit
```

---

## Production Deployment

### 1. Use Environment Variables for Secrets

```bash
# Create .env file (never commit to git)
cp .env.example .env
# Edit .env with production values
```

```yaml
# docker-compose.yml
environment:
  - RCON_PASSWORD=${RCON_PASSWORD}
  - LOG_LEVEL=${LOG_LEVEL:-info}
```

```bash
# Run with env file
docker-compose --env-file .env up -d
```

### 2. Reverse Proxy Setup (Nginx)

```nginx
server {
    listen 80;
    server_name tracker.example.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name tracker.example.com;

    ssl_certificate /etc/letsencrypt/live/tracker.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tracker.example.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        proxy_pass http://localhost:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_read_timeout 86400;
    }
}
```

### 3. Backup Strategy

```bash
#!/bin/bash
# backup.sh - Daily backup script

BACKUP_DIR="/mnt/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup PocketBase data
docker-compose exec -T sandstorm-tracker tar czf - pb_data | \
    gzip > "$BACKUP_DIR/sandstorm-tracker_$DATE.tar.gz"

# Keep only last 7 days
find "$BACKUP_DIR" -name "sandstorm-tracker_*.tar.gz" -mtime +7 -delete

echo "Backup completed: sandstorm-tracker_$DATE.tar.gz"
```

Schedule with cron:

```cron
0 2 * * * /home/tracker/backup.sh >> /var/log/sandstorm-backup.log 2>&1
```

### 4. Monitoring with Prometheus

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "sandstorm-tracker"
    static_configs:
      - targets: ["localhost:8090"]
    metrics_path: "/metrics"
```

---

## Development vs Production

### Development (docker-compose.yml)

```yaml
environment:
  - DEBUG=true
  - LOG_LEVEL=debug
ports:
  - "8090:8090" # Exposed to all interfaces
```

### Production

```yaml
environment:
  - DEBUG=false
  - LOG_LEVEL=info
ports:
  - "127.0.0.1:8090:8090" # Exposed only to localhost
networks:
  - internal # Separate from public network
```

---

## Common Commands

```bash
# Build image
docker build -t sandstorm-tracker:latest .

# Run container
docker run -p 8090:8090 -v pb_data:/app/pb_data sandstorm-tracker:latest

# View logs
docker-compose logs -f --tail=100 sandstorm-tracker

# Execute command in container
docker-compose exec sandstorm-tracker ./sandstorm-tracker --help

# Enter container shell
docker-compose exec sandstorm-tracker sh

# Restart container
docker-compose restart sandstorm-tracker

# Remove containers and volumes
docker-compose down -v

# Prune unused images/containers
docker system prune -a --volumes

# View container stats
docker stats sandstorm-tracker

# Copy files to/from container
docker cp sandstorm-tracker:/app/pb_data/data.db ./backup/
docker cp ./config.yml sandstorm-tracker:/app/
```

---

## Verification Checklist

- [ ] Dockerfile builds without errors
- [ ] Container starts and passes health checks
- [ ] Admin UI accessible at http://localhost:8090/\_/
- [ ] PocketBase collections visible
- [ ] Log files being created
- [ ] Config file mounted correctly
- [ ] Volumes persist data across restarts
- [ ] RCON connectivity working
- [ ] Game logs being watched (if mounted)
- [ ] Archive cron job scheduled
- [ ] Score updates working (if servers configured)

---

## Next Steps

1. **Test locally** with `docker-compose up`
2. **Verify connectivity** to game servers
3. **Set resource limits** appropriate for your hardware
4. **Configure backups** for production
5. **Set up reverse proxy** for HTTPS
6. **Monitor logs** for errors
7. **Plan for scaling** with Docker Swarm or Kubernetes
