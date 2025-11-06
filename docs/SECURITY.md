# Security Best Practices for Sandstorm Tracker

## Overview

Sandstorm Tracker is distributed as a binary that accepts a user-provided configuration file. Your configuration file contains sensitive information (RCON passwords) and should be kept secure.

## Securing Your Configuration File

### 1. File Permissions (Required)

**Linux/Mac:**

```bash
# Create your config from the example
cp sandstorm-tracker.example.yml my-config.yml

# Edit with your actual passwords
nano my-config.yml

# Lock down permissions (owner read/write only)
chmod 600 my-config.yml

# Verify
ls -l my-config.yml
# Should show: -rw------- 1 user user ...
```

**Windows:**

```powershell
# Copy example config
Copy-Item sandstorm-tracker.example.yml my-config.yml

# Edit with your passwords
notepad my-config.yml

# Remove all permissions except owner
icacls my-config.yml /inheritance:r
icacls my-config.yml /grant:r "${env:USERNAME}:(F)"

# Verify
icacls my-config.yml
```

### 2. Use Environment Variables (Optional)

Instead of hardcoding passwords in YAML, use environment variable substitution:

```yaml
servers:
  - name: "Main Server"
    rconPassword: "${RCON_PASSWORD_MAIN}"
```

**Set the variable before running:**

Linux/Mac:

```bash
export RCON_PASSWORD_MAIN="your_actual_password"
./sandstorm-tracker -c my-config.yml
```

Windows:

```powershell
$env:RCON_PASSWORD_MAIN = "your_actual_password"
.\sandstorm-tracker.exe -c my-config.yml
```

## Security Checklist

- [ ] Config file has restrictive permissions (600 on Unix, restricted ACL on Windows)
- [ ] Config file is stored in a secure location (not web-accessible)
- [ ] RCON passwords use strong values (16+ characters, mixed case, numbers, symbols)
- [ ] Different passwords for each server
- [ ] Regular password rotation schedule
- [ ] Server logs are monitored for unauthorized access attempts
- [ ] Database file has appropriate permissions

## Password Requirements

**Strong RCON Password Example:**

- ❌ Weak: `password123`, `admin`, `rcon`
- ✅ Strong: `Kp9#mN$vX2@qL5zR`, `T7r!8nQ$pW3@hM6`

Use a password generator or password manager for best security.

## Deployment Security

### Running as a Service (Linux - systemd)

Create a service file with appropriate permissions:

```ini
# /etc/systemd/system/sandstorm-tracker.service
[Unit]
Description=Sandstorm Tracker
After=network.target

[Service]
Type=simple
User=sandstorm
Group=sandstorm
WorkingDirectory=/opt/sandstorm-tracker
ExecStart=/opt/sandstorm-tracker/sandstorm-tracker -c /opt/sandstorm-tracker/config.yml
Restart=on-failure

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/sandstorm-tracker

[Install]
WantedBy=multi-user.target
```

```bash
# Set up directories with proper permissions
sudo mkdir -p /opt/sandstorm-tracker
sudo useradd -r -s /bin/false sandstorm
sudo chown -R sandstorm:sandstorm /opt/sandstorm-tracker
sudo chmod 600 /opt/sandstorm-tracker/config.yml

# Enable and start
sudo systemctl enable sandstorm-tracker
sudo systemctl start sandstorm-tracker
```

### Running as a Service (Windows)

Use NSSM or Windows Service wrapper:

```powershell
# Install NSSM
choco install nssm

# Create service
nssm install SandstormTracker "C:\sandstorm-tracker\sandstorm-tracker.exe"
nssm set SandstormTracker AppParameters "-c C:\sandstorm-tracker\config.yml"
nssm set SandstormTracker AppDirectory "C:\sandstorm-tracker"

# Start service
nssm start SandstormTracker
```

### Docker Deployment

```dockerfile
FROM scratch
COPY sandstorm-tracker /sandstorm-tracker
ENTRYPOINT ["/sandstorm-tracker"]
```

```bash
# Run with config mounted as volume
docker run -d \
  --name sandstorm-tracker \
  -v /path/to/config.yml:/config.yml:ro \
  -v /path/to/logs:/logs:ro \
  -v /path/to/db:/data \
  sandstorm-tracker:latest -c /config.yml
```

**Security:** Config file is mounted read-only (`:ro`)

## What to Protect

### High Priority (Contains Secrets)

- ✅ Your actual config file (`my-config.yml`)
- ✅ Database file (contains player data)
- ✅ Backup files

### Safe to Share

- ✅ The binary (`sandstorm-tracker` / `sandstorm-tracker.exe`)
- ✅ Example config file (`sandstorm-tracker.example.yml`)
- ✅ Documentation files
