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

## Configuration Example

```yaml
# Optional global settings
listen: ":9100"              # HTTP listen address (can be overridden by -listen flag)
http_auth:                   # Optional HTTP basic authentication
  username: "admin"
  password: "secret"

hosts:
  - host: "192.168.1.100"
    user: "monitoring"
    password: "your_password"  # Use either password or private_key
    # private_key: "/path/to/id_rsa"  # SSH private key (alternative to password)
    port: 22
    monitors:
      stat: true
      processes:
        - path_pattern: "/proc/[0-9]*/cmdline"
          patterns: ["nginx", "java"]
      files:
        - path: "/var/log/"
```

## Security Notes

- **SSH Host Keys**: The exporter uses `InsecureIgnoreHostKey()` and will trust all SSH host keys automatically
- **Passwords**: Stored in plaintext in config file - protect with `chmod 600 config.yaml`
- **SSH Authentication**: Supports both password and private key authentication
- **HTTP Authentication**: Optional HTTP basic auth to protect metrics endpoint

## Metrics

See full documentation for complete metrics list.

## License

MIT License - see LICENSE file for details.
