# SSH Exporter

[English](README.md) | 简体中文

一个用 Go 编写的 Prometheus 导出器，通过 SSH 从远程 Linux 主机收集指标。无需在目标主机上安装任何代理，即可监控进程、文件和系统统计信息（CPU、内存、磁盘）。

## 特性

- **进程监控**：统计匹配特定模式的进程数量
- **文件监控**：跟踪文件大小、年龄和修改时间戳
- **系统统计**：
  - CPU 使用率、空闲时间、I/O 等待、上下文切换、中断
  - 内存使用、可用内存、缓冲区、缓存
  - 每个分区的磁盘使用情况
- **多主机支持**：通过 goroutine 并发监控多个主机
- **基于 SSH**：目标主机无需安装代理，仅需 SSH 访问
- **Prometheus 原生**：实现标准的 Prometheus Collector 接口
- **采集器跨平台**：采集器可运行于 Linux、Windows、macOS（AMD64/ARM64）
- **目标系统要求**：仅支持监控 Linux 主机（依赖 `/proc` 文件系统）

## 快速开始

### 安装

从 [Releases](https://github.com/5iop/ssh_exporter/releases) 页面下载最新版本。

### 配置

1. 复制示例配置文件：

```bash
cp config.yaml.example config.yaml
```

2. 编辑 `config.yaml` 填写主机信息：

```yaml
# 可选的全局配置
listen: ":9100"              # HTTP监听地址（可被 -listen 参数覆盖）
http_auth:                   # 可选的HTTP基本认证
  username: "admin"
  password: "secret"

hosts:
  - host: "192.168.1.100"
    user: "monitoring"
    password: "your_password"  # 使用密码或私钥二选一
    # private_key: "/path/to/id_rsa"  # SSH私钥路径（密码的替代方案）
    port: 22
    monitors:
      stat: true
      processes:
        - path_pattern: "/proc/[0-9]*/cmdline"
          patterns: ["nginx", "java"]
      files:
        - path: "/var/log/"
```

3. 保护配置文件：

```bash
chmod 600 config.yaml
```

### 运行

```bash
# 使用默认配置（config.yaml）和端口（:9100）
./ssh_exporter

# 自定义配置文件和端口
./ssh_exporter -config /path/to/config.yaml -listen :8080
```

### Prometheus 配置

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'ssh_exporter'
    static_configs:
      - targets: ['localhost:9100']
    scrape_interval: 30s
    scrape_timeout: 25s
```

## 指标说明

完整的指标列表请参考英文文档。主要包括：

### 主机指标
- `host_ssh_status` - SSH 连接状态
- `host_last_check_timestamp` - 最后检查时间

### 进程指标
- `process_pattern_count` - 匹配模式的进程数

### 文件指标
- `file_size_bytes` - 文件大小
- `file_last_modified_timestamp` - 最后修改时间
- `file_age_minutes` - 文件年龄（分钟）

### 系统统计（stat: true）

#### CPU 指标
- `cpu_user_seconds_total` - 用户态 CPU 时间
- `cpu_system_seconds_total` - 内核态 CPU 时间
- `cpu_idle_seconds_total` - CPU 空闲时间
- `cpu_iowait_seconds_total` - I/O 等待时间
- `cpu_usage_percent` - CPU 使用率（0-1）
- `context_switches_total` - 上下文切换次数
- `interrupts_total` - 中断次数
- `processes_running` - 运行中的进程数
- `processes_blocked` - 阻塞的进程数

#### 内存指标
- `memory_total_bytes` - 总内存
- `memory_free_bytes` - 空闲内存
- `memory_available_bytes` - 可用内存
- `memory_buffers_bytes` - 缓冲区内存
- `memory_cached_bytes` - 缓存内存
- `memory_usage_percent` - 内存使用率（0-1）

#### 磁盘指标
- `disk_total_bytes` - 总磁盘空间
- `disk_used_bytes` - 已用空间
- `disk_free_bytes` - 可用空间
- `disk_usage_percent` - 磁盘使用率（0-100）

## 性能考虑

- **每次抓取的 SSH 命令数（启用 stat）**：4 条命令
  - 2x `cat /proc/stat`（间隔 1 秒用于 CPU 使用率计算）
  - 1x `cat /proc/meminfo`
  - 1x `df -B1 ...`
  
- **抓取耗时**：每个主机约 1.2 秒（主要是 CPU 采样间隔）
- **并发收集**：所有主机通过 goroutine 并发监控
- **推荐的 Prometheus 抓取间隔**：15-60 秒

## SSH 要求

SSH 用户必须具有以下权限：

- 读取 `/proc/*/cmdline`（用于进程监控）
- 读取 `/proc/stat` 和 `/proc/meminfo`（用于系统统计）
- 执行 `df` 命令的权限（用于磁盘监控）
- 读取被监控文件路径的权限（用于文件监控）

## 安全注意事项

- **SSH 主机密钥**：采集器使用 `InsecureIgnoreHostKey()` 会自动信任所有 SSH 主机密钥
- **密码存储**：密码以明文形式存储在配置文件中，请使用 `chmod 600 config.yaml` 保护
- **SSH 认证**：支持密码认证和私钥认证两种方式
- **HTTP 认证**：可选的 HTTP 基本认证保护指标端点

## Grafana 查询示例

**CPU 使用率百分比：**
```promql
cpu_usage_percent{host="192.168.1.100"} * 100
```

**内存使用率百分比：**
```promql
(1 - (memory_available_bytes / memory_total_bytes)) * 100
```

**磁盘使用率：**
```promql
disk_usage_percent{host="192.168.1.100", mount_point="/"}
```

**进程计数：**
```promql
process_pattern_count{host="192.168.1.100", pattern="nginx"}
```

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
