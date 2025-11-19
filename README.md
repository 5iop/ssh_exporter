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

Download from releases or build from source:

```bash
git clone https://github.com/5iop/ssh_exporter.git
cd ssh_exporter
go build -o ssh_exporter .
```

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
hosts:
  - host: "192.168.1.100"
    user: "monitoring"
    password: "your_password"
    port: 22
    monitors:
      stat: true
      processes:
        - path_pattern: "/proc/[0-9]*/cmdline"
          patterns: ["nginx", "java"]
      files:
        - path: "/var/log/"
```

## Metrics

See full documentation for complete metrics list.

## Building

```bash
make build              # Current platform
make build-linux-amd64  # Linux AMD64
make build-all          # All platforms
```

## License

MIT License - see LICENSE file for details.
