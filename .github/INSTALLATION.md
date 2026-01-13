# Installation Guide

## Quick Start

### 1. Download Binary

```bash
curl -L -o /usr/local/bin/sentinel https://github.com/yourusername/sentinel/releases/download/v1.0.0/sentinel

# Make executable
sudo chmod +x /usr/local/bin/sentinel

# Verify
sentinel --help
```

### 2. Setup Directories

```bash
sudo mkdir -p /etc/sentinel
sudo mkdir -p /var/run/sentinel
sudo mkdir -p /var/log/sentinel
sudo chown root:root /var/log/sentinel
```

### 3. First Run (Auto-creates Config)

```bash
# Run once to generate default config
sudo /usr/local/bin/sentinel --log-level debug

# Ctrl+C to stop, then edit config
sudo nano /etc/sentinel/config.toml

# Secure config (contains API keys)
sudo chmod 600 /etc/sentinel/config.toml
```

### 4. Copy YARA Signatures

```bash
# Copy signature files (from repo)
sudo cp -r signatures /etc/sentinel/
```

### 5. Install Systemd Service

```bash
# Copy service file (from repo)
sudo cp sentinel.service /etc/systemd/system/

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

### 6. Verify Installation

```bash
# Check status
sudo systemctl status sentinel

# View logs
sudo journalctl -u sentinel -f

# Check if running
sentinel --action status
```

## Configuration

Edit `/etc/sentinel/config.toml`:

```toml
# Monitoring paths (can have multiple)
watchdogPath = [
    "/var/lib/pterodactyl/volumes",
    "/var/www/html"
]

# YARA rules location
SignaturePath = "/etc/sentinel/signatures"

# Max file size to scan
maxFileSizeMB = 500

# AI Analysis (Ollama or Groq)
[INTEGRATION.AI]
enabled = true
generate_endpoint = "http://localhost:11434/api/generate"
use_groq = false

# Discord Notifications
[INTEGRATION.DISCORD]
enabled = true
webhook_url = "https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN"

# Pterodactyl Auto-Suspend
[PLUGINS.PterodactylAutoSuspend]
hostname = "https://panel.example.com"
api_key = "ptla_xxxxxxxxxxxx"
```

## Uninstall

```bash
sudo systemctl stop sentinel
sudo systemctl disable sentinel
sudo rm /etc/systemd/system/sentinel.service
sudo rm /usr/local/bin/sentinel
sudo rm -rf /etc/sentinel
sudo systemctl daemon-reload
```

## Troubleshooting

### Permission Denied

```bash
sudo chmod +x /usr/local/bin/sentinel
```

### Can't Find YARA Rules

```bash
ls -la /etc/sentinel/signatures/
# Should have .yar or .yara files
```

### Log File Errors

```bash
sudo chown root:root /var/log/sentinel
sudo chmod 755 /var/log/sentinel
```

### View Debug Logs

```bash
sudo systemctl stop sentinel
/usr/local/bin/sentinel --log-level debug
```

## Updating

```bash
# Download new binary
curl -L -o /usr/local/bin/sentinel https://github.com/yourusername/sentinel/releases/download/vX.X.X/sentinel-linux-amd64

# Make executable
sudo chmod +x /usr/local/bin/sentinel

# Restart service
sudo systemctl restart sentinel
```
