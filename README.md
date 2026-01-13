# Sentinel by Novel

Advanced Abuse Detection & Prevention System for Pterodactyl

## Overview

Sentinel is a high-performance, minimal-overhead abuse detection system designed for monitoring Pterodactyl servers and Docker containers. It detects malicious activities including NEZHA proxies, crypto miners, shell access, and other abuse patterns.

## Features

- **High Performance**: Optimized for 6000-8000 files/sec processing with auto-tuning
- **Minimal Resource Usage**: Efficient goroutine pools and memory management
- **Daemon Support**: Native systemd service and binary daemon management
- **YARA Integration**: Real-time file scanning with customizable rules (including NEZHA detection)
- **Plugin System**: Extensible architecture for custom actions (Pterodactyl Auto-Suspend)
- **AI Analysis**: Groq/OpenAI or Ollama integration for abuse scoring
- **Discord Webhooks**: Real-time notifications for flagged content
- **Auto-Suspend**: Automatic Pterodactyl server suspension on detection

## Installation

### Requirements

- Go 1.21+ (for building from source)
- YARA 4.3+ library
- Linux/Unix system
- Root or sudo access

### Quick Install

```bash
# Create config directory
sudo mkdir -p /etc/sentinel

# Download binary (auto-detects architecture)
curl -L -o /usr/local/bin/sentinel "https://github.com/your-org/sentinel/releases/latest/download/sentinel_linux_$([[ "$(uname -m)" == "x86_64" ]] && echo "amd64" || echo "arm64")"
sudo chmod u+x /usr/local/bin/sentinel

# Verify installation
sentinel --help
```

### Setup Configuration

```bash
# Copy systemd service
sudo mkdir -p /etc/systemd/system
sudo cp sentinel.service /etc/systemd/system/sentinel.service

# Copy configuration and rules
sudo cp config.toml /etc/sentinel/
sudo cp -r signatures /etc/sentinel/

# Edit configuration for your environment
sudo nano /etc/sentinel/config.toml

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

### Build from Source

For developers or custom builds:

```bash
cd sentinel
make build
```

For WSL users with CGO issues:

```bash
CGO_LDFLAGS="-L/usr/lib" make build
```

## Usage

### Binary Commands

```bash
# Run in foreground
sentinel

# Daemon management
sentinel --action start
sentinel --action stop
sentinel --action restart
sentinel --action status

# Custom config and log level
sentinel --config /etc/sentinel/config.toml --log-level debug

# Systemd
sudo systemctl start sentinel
sudo systemctl stop sentinel
sudo systemctl restart sentinel
sudo systemctl status sentinel
sudo journalctl -u sentinel -f
```

### Configuration

Edit `/etc/sentinel/config.toml`:

- **watchdogPath**: Directories to monitor
- **SignaturePath**: Path to YARA rules (`/etc/sentinel/signatures`)
- **maxFileSizeMB**: Maximum file size to scan (default: 500)
- **INTEGRATION.AI**: Enable/disable AI analysis
- **INTEGRATION.DISCORD**: Discord webhook notifications
- **PLUGINS.PterodactylAutoSuspend**: Auto-suspension on detection

## Performance Tuning

The system auto-tunes based on:

- CPU cores (worker pool = CPU × 2)
- Available memory (buffer size scaling)
- File system capabilities

For optimal performance:

```bash
# Increase inotify watches (Linux)
sudo sysctl -w fs.inotify.max_user_watches=100000
echo 'fs.inotify.max_user_watches=100000' | sudo tee -a /etc/sysctl.conf

# Monitor performance
sudo journalctl -u sentinel -f --output cat | grep -E "workers|tuned"

# View debug logs
sentinel --log-level debug
```

## Architecture

- **Watcher**: fsnotify-based file monitoring with batched events
- **Scanner**: Pre-compiled YARA rules for fast scanning
- **Worker Pool**: Configurable goroutines for parallel processing
- **Plugins**: Interface-based extensibility
- **Daemon**: Native process management with PID files

## Security

- Runs as root for file system access
- Minimal dependencies
- Memory limits on file scanning
- Secure API key handling
