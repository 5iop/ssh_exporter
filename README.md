# SSH Exporter

English | [简体中文](README_CN.md)

A Prometheus exporter written in Go that collects metrics from remote Linux hosts via SSH. It monitors processes, files, and system statistics (CPU, memory, disk) without requiring any agent installation on target hosts.

## Features

- **Process Monitoring**: Count processes matching specific patterns
- **File Monitoring**: Track file size, age, and modification timestamps
- **System Statistics**: CPU, memory, disk usage
- **Multi-Host Support**: Monitor multiple hosts concurrently
- **SSH-Based**: No agent required on target hosts
- **Prometheus Native**: Standard Prometheus Collector interface
- **Cross-Platform Exporter**: Runs on Linux, Windows, macOS (AMD64/ARM64)
- **Linux Target Systems**: Monitors Linux hosts via SSH (requires `/proc` filesystem)

## Quick Start

### Installation

Download the latest release from [Releases](https://github.com/5iop/ssh_exporter/releases) page.

### Configuration

Copy and edit the example config:

```bash
cp config.yaml.example config.yaml
chmod 600 config.yaml
```

### Running

```bash
./ssh_exporter
./ssh_exporter -config /path/to/config.yaml -listen :8080
```

## Configuration

### Global Settings

```yaml
# Optional global settings
listen: ":9100"              # HTTP listen address (default: :9109)
metric_prefix: "ssh_"        # Prefix for all metrics (default: no prefix)
http_auth:                   # Optional HTTP basic authentication
  username: "admin"
  password: "secret"
```

**Global Options:**
- `listen` - HTTP server listen address (can be overridden by `-listen` command-line flag)
- `metric_prefix` - Optional prefix added to all metric names (e.g., `ssh_cpu_usage_percent`)
- `http_auth` - Optional HTTP basic authentication to protect metrics endpoint

### Host Configuration

Each host can have different monitoring configurations:

```yaml
hosts:
  - host: "192.168.1.100"        # IP address or hostname
    user: "monitoring"            # SSH username
    password: "your_password"     # SSH password (use password OR private_key)
    private_key: "/path/to/id_rsa"  # SSH private key path (alternative to password)
    port: 22                      # SSH port (default: 22)

    monitors:
      # System statistics (CPU, memory, disk)
      stat: true

      # Process monitoring
      processes:
        - patterns: ["nginx", "java", "python"]  # Process names to count

      # File monitoring
      files:
        - path: "/var/log/"       # Directory to monitor
          labels:
            - pattern: ".*\\.log$"
              name: "type"
              value: "logfile"
```

**Host Options:**
- `host` - Target hostname or IP address (required)
- `user` - SSH username (required)
- `password` - SSH password (optional, use password OR private_key)
- `private_key` - Path to SSH private key file (optional, alternative to password)
- `port` - SSH port number (optional, default: 22)

**Monitor Types:**
- `stat` - Collect system statistics (CPU, memory, disk usage)
- `processes` - Count processes by name pattern
- `files` - Monitor file size, age, and modification time

## Security Notes

- **SSH Host Keys**: The exporter uses `InsecureIgnoreHostKey()` and will trust all SSH host keys automatically
- **Passwords**: Stored in plaintext in config file - protect with `chmod 600 config.yaml`
- **SSH Authentication**: Supports both password and private key authentication
- **HTTP Authentication**: Optional HTTP basic auth to protect metrics endpoint

## Metrics

See full documentation for complete metrics list.

## License

Apache License 2.0 - see LICENSE file for details.
