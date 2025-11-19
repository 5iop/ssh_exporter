package collector

import (
	"log"
	"os"
	"sync"
	"time"

	"ssh_exporter/config"
	sshclient "ssh_exporter/ssh"

	"github.com/prometheus/client_golang/prometheus"
)

var logger = log.New(os.Stdout, "[Collector] ", log.LstdFlags)

// SSHCollector 实现Prometheus Collector接口
type SSHCollector struct {
	config *config.Config
	mu     sync.Mutex

	// 进程监控指标
	processPatternCount *prometheus.Desc

	// 文件监控指标
	fileSize         *prometheus.Desc
	fileLastModified *prometheus.Desc
	fileAgeMinutes   *prometheus.Desc

	// 主机状态指标
	hostSSHStatus *prometheus.Desc
	hostLastCheck *prometheus.Desc

	// CPU指标
	cpuUserSeconds   *prometheus.Desc
	cpuSystemSeconds *prometheus.Desc
	cpuIdleSeconds   *prometheus.Desc
	cpuIowaitSeconds *prometheus.Desc
	cpuUsagePercent  *prometheus.Desc
	contextSwitches  *prometheus.Desc
	interrupts       *prometheus.Desc
	processesRunning *prometheus.Desc
	processesBlocked *prometheus.Desc

	// 内存指标
	memoryTotalBytes     *prometheus.Desc
	memoryFreeBytes      *prometheus.Desc
	memoryAvailableBytes *prometheus.Desc
	memoryBuffersBytes   *prometheus.Desc
	memoryCachedBytes    *prometheus.Desc
	memoryUsagePercent   *prometheus.Desc

	// 磁盘指标
	diskTotalBytes   *prometheus.Desc
	diskUsedBytes    *prometheus.Desc
	diskFreeBytes    *prometheus.Desc
	diskUsagePercent *prometheus.Desc
}

// NewSSHCollector 创建新的SSH Collector
func NewSSHCollector(cfg *config.Config) *SSHCollector {
	return &SSHCollector{
		config: cfg,
		processPatternCount: prometheus.NewDesc(
			"process_pattern_count",
			"Count of pattern in process cmdlines",
			[]string{"host", "pattern"},
			nil,
		),
		fileSize: prometheus.NewDesc(
			"file_size_bytes",
			"File size in bytes",
			[]string{"host", "path", "filename"},
			nil,
		),
		fileLastModified: prometheus.NewDesc(
			"file_last_modified_timestamp",
			"Last modified timestamp of file",
			[]string{"host", "path", "filename"},
			nil,
		),
		fileAgeMinutes: prometheus.NewDesc(
			"file_age_minutes",
			"Minutes since last modification",
			[]string{"host", "path", "filename"},
			nil,
		),
		hostSSHStatus: prometheus.NewDesc(
			"host_ssh_status",
			"SSH connection status to host (1: success, 0: failure)",
			[]string{"host"},
			nil,
		),
		hostLastCheck: prometheus.NewDesc(
			"host_last_check_timestamp",
			"Last successful check timestamp of host",
			[]string{"host"},
			nil,
		),
		cpuUserSeconds: prometheus.NewDesc(
			"cpu_user_seconds_total",
			"Total CPU time spent in user mode",
			[]string{"host"},
			nil,
		),
		cpuSystemSeconds: prometheus.NewDesc(
			"cpu_system_seconds_total",
			"Total CPU time spent in system mode",
			[]string{"host"},
			nil,
		),
		cpuIdleSeconds: prometheus.NewDesc(
			"cpu_idle_seconds_total",
			"Total CPU idle time",
			[]string{"host"},
			nil,
		),
		cpuIowaitSeconds: prometheus.NewDesc(
			"cpu_iowait_seconds_total",
			"Total CPU time waiting for I/O",
			[]string{"host"},
			nil,
		),
		cpuUsagePercent: prometheus.NewDesc(
			"cpu_usage_percent",
			"CPU usage percentage",
			[]string{"host"},
			nil,
		),
		contextSwitches: prometheus.NewDesc(
			"context_switches_total",
			"Total number of context switches",
			[]string{"host"},
			nil,
		),
		interrupts: prometheus.NewDesc(
			"interrupts_total",
			"Total number of interrupts",
			[]string{"host"},
			nil,
		),
		processesRunning: prometheus.NewDesc(
			"processes_running",
			"Number of processes in running state",
			[]string{"host"},
			nil,
		),
		processesBlocked: prometheus.NewDesc(
			"processes_blocked",
			"Number of processes blocked waiting for I/O",
			[]string{"host"},
			nil,
		),
		memoryTotalBytes: prometheus.NewDesc(
			"memory_total_bytes",
			"Total memory in bytes",
			[]string{"host"},
			nil,
		),
		memoryFreeBytes: prometheus.NewDesc(
			"memory_free_bytes",
			"Free memory in bytes",
			[]string{"host"},
			nil,
		),
		memoryAvailableBytes: prometheus.NewDesc(
			"memory_available_bytes",
			"Available memory in bytes",
			[]string{"host"},
			nil,
		),
		memoryBuffersBytes: prometheus.NewDesc(
			"memory_buffers_bytes",
			"Memory used for buffers in bytes",
			[]string{"host"},
			nil,
		),
		memoryCachedBytes: prometheus.NewDesc(
			"memory_cached_bytes",
			"Memory used for cache in bytes",
			[]string{"host"},
			nil,
		),
		memoryUsagePercent: prometheus.NewDesc(
			"memory_usage_percent",
			"Memory usage percentage",
			[]string{"host"},
			nil,
		),
		diskTotalBytes: prometheus.NewDesc(
			"disk_total_bytes",
			"Total disk space in bytes",
			[]string{"host", "device", "mount_point"},
			nil,
		),
		diskUsedBytes: prometheus.NewDesc(
			"disk_used_bytes",
			"Used disk space in bytes",
			[]string{"host", "device", "mount_point"},
			nil,
		),
		diskFreeBytes: prometheus.NewDesc(
			"disk_free_bytes",
			"Free disk space in bytes",
			[]string{"host", "device", "mount_point"},
			nil,
		),
		diskUsagePercent: prometheus.NewDesc(
			"disk_usage_percent",
			"Disk usage percentage",
			[]string{"host", "device", "mount_point"},
			nil,
		),
	}
}

// Describe 实现Prometheus Collector接口
func (c *SSHCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.processPatternCount
	ch <- c.fileSize
	ch <- c.fileLastModified
	ch <- c.fileAgeMinutes
	ch <- c.hostSSHStatus
	ch <- c.hostLastCheck
	ch <- c.cpuUserSeconds
	ch <- c.cpuSystemSeconds
	ch <- c.cpuIdleSeconds
	ch <- c.cpuIowaitSeconds
	ch <- c.cpuUsagePercent
	ch <- c.contextSwitches
	ch <- c.interrupts
	ch <- c.processesRunning
	ch <- c.processesBlocked
	ch <- c.memoryTotalBytes
	ch <- c.memoryFreeBytes
	ch <- c.memoryAvailableBytes
	ch <- c.memoryBuffersBytes
	ch <- c.memoryCachedBytes
	ch <- c.memoryUsagePercent
	ch <- c.diskTotalBytes
	ch <- c.diskUsedBytes
	ch <- c.diskFreeBytes
	ch <- c.diskUsagePercent
}

// Collect 实现Prometheus Collector接口
func (c *SSHCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var wg sync.WaitGroup
	metricsChan := make(chan prometheus.Metric, 100)

	// 为每个主机启动一个goroutine
	for _, hostConfig := range c.config.Hosts {
		wg.Add(1)
		go func(hc config.HostConfig) {
			defer wg.Done()
			c.collectHostMetrics(hc, metricsChan)
		}(hostConfig)
	}

	// 等待所有goroutine完成并关闭channel
	go func() {
		wg.Wait()
		close(metricsChan)
	}()

	// 收集所有指标
	for metric := range metricsChan {
		ch <- metric
	}
}

// collectHostMetrics 收集单个主机的指标
func (c *SSHCollector) collectHostMetrics(hostConfig config.HostConfig, ch chan<- prometheus.Metric) {
	logger.Printf("Collecting metrics for host: %s", hostConfig.Host)
	currentTime := float64(time.Now().Unix())

	// 创建SSH客户端
	client := sshclient.NewClient(hostConfig.Host, hostConfig.User, hostConfig.Password, hostConfig.Port)
	err := client.Connect()
	if err != nil {
		logger.Printf("Failed to connect to %s: %v", hostConfig.Host, err)
		// 报告连接失败
		ch <- prometheus.MustNewConstMetric(
			c.hostSSHStatus,
			prometheus.GaugeValue,
			0,
			hostConfig.Host,
		)
		return
	}
	defer client.Close()

	// 报告连接成功
	ch <- prometheus.MustNewConstMetric(
		c.hostSSHStatus,
		prometheus.GaugeValue,
		1,
		hostConfig.Host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hostLastCheck,
		prometheus.GaugeValue,
		currentTime,
		hostConfig.Host,
	)

	// 收集进程监控指标
	for _, processMonitor := range hostConfig.Monitors.Processes {
		c.collectProcessMetrics(client, hostConfig.Host, processMonitor, ch)
	}

	// 收集文件监控指标
	for _, fileMonitor := range hostConfig.Monitors.Files {
		c.collectFileMetrics(client, hostConfig.Host, fileMonitor, ch, currentTime)
	}

	// 收集系统统计指标
	if hostConfig.Monitors.Stat {
		c.collectStatMetrics(client, hostConfig.Host, ch, currentTime)
	}
}
