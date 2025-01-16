# SMTP Relay Server

A lightweight SMTP relay server written in Go, designed for efficient email delivery with configurable rate limiting and TLS support.

## Features

- Supports multiple SMTP protocols:
  - SMTPS/TLS (port 465)
  - STARTTLS (port 587) 
  - Unencrypted SMTP (port 25)
- Configurable encryption per listener
- Configurable rate limiting
- IP blocking
- Multi-platform support (Windows, Linux, MacOS)
- Detailed logging
- Automatic service installation
- Comprehensive service management
- Log rotation and monitoring

## Installation

### Prerequisites

- Go 1.20+ installed
- Git installed
- System administrator privileges

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/WikiHacker/go-relay-server.git
cd go-relay-server
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the server by editing `config/config.json`

4. Compile the server:
```bash
# Linux/MacOS
chmod +x script/compile.sh
./script/compile.sh

# Windows
./script/compile.ps1
```

5. Install and start the service:
```bash
# Linux/MacOS
sudo ./script/setup-installation.sh

# Windows (Run as Administrator)
.\script\setup-installation.ps1
```

## Configuration

Edit `config/config.json` with your desired settings:

```json
{
  "Listeners": [
    {
      "host": "0.0.0.0",
      "port": "25",
      "encryption": "none",
      "require_auth": false
    },
    {
      "host": "0.0.0.0", 
      "port": "465",
      "encryption": "tls",
      "require_auth": true
    },
    {
      "host": "0.0.0.0",
      "port": "587",
      "encryption": "starttls",
      "require_auth": true
    }
  ],
  "DefaultRelay": "smtp.example.com:25",
  "TLSCertFile": "certs/cert.pem",
  "TLSKeyFile": "certs/key.pem",
  "RateLimiting": {
    "RequestsPerMinute": 100,
    "BurstLimit": 20,
    "ExemptIPs": ["127.0.0.1"]
  },
  "BlockList": ["spamdomain.com"]
}
```

## Service Management

The server provides comprehensive service management through the `manage-service.sh` script:

### Installation
```bash
sudo ./script/manage-service.sh install [install-dir]
```

### Common Operations
- Start service: `sudo ./script/manage-service.sh start`
- Stop service: `sudo ./script/manage-service.sh stop`
- Restart service: `sudo ./script/manage-service.sh restart`
- Check status: `sudo ./script/manage-service.sh status`
- View logs: `sudo ./script/manage-service.sh logs`
- Uninstall service: `sudo ./script/manage-service.sh uninstall`

### Windows Specific
Run all commands from an elevated PowerShell prompt:
```powershell
# Install service
.\script\manage-service.ps1 install

# Start service
.\script\manage-service.ps1 start

# View logs
.\script\manage-service.ps1 logs
```

## Directory Structure

The server requires the following directory structure:

```
.
├── build/               # Compiled binaries
├── config/              # Configuration files
│   ├── config.json      # Main configuration
│   └── certs/           # TLS certificates (if using TLS)
├── logs/                # Log files (created automatically)
│   ├── output.log       # Standard output
│   └── error.log        # Error output
└── script/              # Management scripts
```

## Advanced Configuration

### Log Rotation
Logs are automatically rotated when they reach 100MB. To configure:

1. Edit `config/config.json`:
```json
{
  "Logging": {
    "MaxSizeMB": 100,
    "MaxBackups": 5,
    "MaxAgeDays": 30
  }
}
```

### Rate Limiting
Configure rate limiting in `config/config.json`:
```json
{
  "RateLimiting": {
    "RequestsPerMinute": 100,
    "BurstLimit": 20,
    "ExemptIPs": ["127.0.0.1"]
  }
}
```

### TLS Configuration
To enable TLS, provide certificate files in `config/certs/` and update:
```json
{
  "Listeners": [
    {
      "host": "0.0.0.0",
      "port": "465",
      "encryption": "tls",
      "require_auth": true
    }
  ],
  "TLSCertFile": "certs/cert.pem",
  "TLSKeyFile": "certs/key.pem"
}
```

## Troubleshooting

### Common Issues

1. **Service fails to start**
   - Verify configuration file exists and is valid
   - Check logs: `sudo ./script/manage-service.sh logs`
   - Ensure required ports are open

2. **Permission denied errors**
   - Run commands with sudo/Administrator privileges
   - Verify installation directory permissions

3. **Connection issues**
   - Verify firewall settings
   - Check network connectivity
   - Validate TLS certificates (if using)

## License

MIT License
