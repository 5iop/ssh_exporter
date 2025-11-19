package collector

import (
	"strconv"
	"strings"
	"time"

	sshclient "ssh_exporter/ssh"

	"github.com/prometheus/client_golang/prometheus"
)

// CPUStats CPU统计信息
type CPUStats struct {
	User         float64
	System       float64
	Idle         float64
	Iowait       float64
	Total        float64 // 总时间 = user + system + idle + iowait + ...
	Ctxt         float64 // 上下文切换
	Intr         float64 // 中断
	ProcsRunning float64
	ProcsBlocked float64
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	Total        float64
	Free         float64
	Available    float64
	Buffers      float64
	Cached       float64
	UsagePercent float64
}

// DiskStats 磁盘统计信息
type DiskStats struct {
	Device       string
	MountPoint   string
	Total        float64
	Used         float64
	Free         float64
	UsagePercent float64
}

// collectStatMetrics 收集系统统计指标 (CPU、内存、磁盘)
func (c *SSHCollector) collectStatMetrics(client *sshclient.Client, host string, ch chan<- prometheus.Metric, currentTime float64) {
	// 收集CPU指标
	c.collectCPUMetrics(client, host, ch)

	// 收集内存指标
	c.collectMemoryMetrics(client, host, ch)

	// 收集磁盘指标
	c.collectDiskMetrics(client, host, ch)
}

// collectCPUMetrics 收集CPU指标
func (c *SSHCollector) collectCPUMetrics(client *sshclient.Client, host string, ch chan<- prometheus.Metric) {
	// 第一次读取 /proc/stat
	output1, err := client.ExecuteCommand("cat /proc/stat")
	if err != nil {
		logger.Printf("Failed to read /proc/stat on %s: %v", host, err)
		return
	}

	stats1 := parseCPUStats(output1)
	if stats1 == nil {
		logger.Printf("Failed to parse CPU stats on %s", host)
		return
	}

	// 发送CPU累计时间指标
	ch <- prometheus.MustNewConstMetric(
		c.cpuUserSeconds,
		prometheus.CounterValue,
		stats1.User,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.cpuSystemSeconds,
		prometheus.CounterValue,
		stats1.System,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.cpuIdleSeconds,
		prometheus.CounterValue,
		stats1.Idle,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.cpuIowaitSeconds,
		prometheus.CounterValue,
		stats1.Iowait,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.contextSwitches,
		prometheus.CounterValue,
		stats1.Ctxt,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.interrupts,
		prometheus.CounterValue,
		stats1.Intr,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processesRunning,
		prometheus.GaugeValue,
		stats1.ProcsRunning,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processesBlocked,
		prometheus.GaugeValue,
		stats1.ProcsBlocked,
		host,
	)

	// 计算CPU使用率：等待1秒后再次采样
	time.Sleep(1 * time.Second)

	output2, err := client.ExecuteCommand("cat /proc/stat")
	if err != nil {
		logger.Printf("Failed to read /proc/stat (2nd time) on %s: %v", host, err)
		// 即使第二次读取失败，也不影响其他指标
	} else {
		stats2 := parseCPUStats(output2)
		if stats2 != nil {
			// 计算差值
			totalDelta := stats2.Total - stats1.Total
			idleDelta := stats2.Idle - stats1.Idle

			if totalDelta > 0 {
				// CPU使用率 = 1 - idle增量 / total增量 (范围: 0-1)
				cpuUsage := 1.0 - idleDelta/totalDelta

				// 边界检查：确保使用率在 0-1 之间
				if cpuUsage < 0 {
					logger.Printf("Warning: CPU usage calculated as %.4f (negative), setting to 0", cpuUsage)
					cpuUsage = 0
				} else if cpuUsage > 1 {
					logger.Printf("Warning: CPU usage calculated as %.4f (>1), clamping to 1", cpuUsage)
					cpuUsage = 1
				}

				ch <- prometheus.MustNewConstMetric(
					c.cpuUsagePercent,
					prometheus.GaugeValue,
					cpuUsage,
					host,
				)

				logger.Printf("Collected CPU metrics for host %s (usage: %.4f, total_delta: %.2f, idle_delta: %.2f)",
					host, cpuUsage, totalDelta, idleDelta)
			} else {
				logger.Printf("Collected CPU metrics for host %s (unable to calculate usage: total_delta=%.2f)", host, totalDelta)
			}
		}
	}
}

// parseCPUStats 解析 /proc/stat 输出
func parseCPUStats(output string) *CPUStats {
	stats := &CPUStats{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "cpu":
			// cpu line: cpu user nice system idle iowait irq softirq steal guest guest_nice
			// Total = user + nice + system + idle + iowait + irq + softirq + steal
			// 注意: guest和guest_nice已经包含在user和nice中，不需要再加
			if len(fields) >= 9 {
				// 时间单位是 USER_HZ (通常是1/100秒), 转换为秒
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				iowait, _ := strconv.ParseFloat(fields[5], 64)
				irq, _ := strconv.ParseFloat(fields[6], 64)
				softirq, _ := strconv.ParseFloat(fields[7], 64)
				steal, _ := strconv.ParseFloat(fields[8], 64)

				stats.User = user / 100.0
				stats.System = system / 100.0
				stats.Idle = idle / 100.0
				stats.Iowait = iowait / 100.0

				// Total = user + nice + system + idle + iowait + irq + softirq + steal
				stats.Total = (user + nice + system + idle + iowait + irq + softirq + steal) / 100.0
			}
		case "ctxt":
			// 上下文切换
			if len(fields) >= 2 {
				stats.Ctxt, _ = strconv.ParseFloat(fields[1], 64)
			}
		case "intr":
			// 中断总数
			if len(fields) >= 2 {
				stats.Intr, _ = strconv.ParseFloat(fields[1], 64)
			}
		case "procs_running":
			// 运行中的进程数
			if len(fields) >= 2 {
				stats.ProcsRunning, _ = strconv.ParseFloat(fields[1], 64)
			}
		case "procs_blocked":
			// 阻塞的进程数
			if len(fields) >= 2 {
				stats.ProcsBlocked, _ = strconv.ParseFloat(fields[1], 64)
			}
		}
	}

	return stats
}

// collectMemoryMetrics 收集内存指标
func (c *SSHCollector) collectMemoryMetrics(client *sshclient.Client, host string, ch chan<- prometheus.Metric) {
	// 读取 /proc/meminfo
	output, err := client.ExecuteCommand("cat /proc/meminfo")
	if err != nil {
		logger.Printf("Failed to read /proc/meminfo on %s: %v", host, err)
		return
	}

	stats := parseMemoryStats(output)
	if stats == nil {
		logger.Printf("Failed to parse memory stats on %s", host)
		return
	}

	// 发送内存指标
	ch <- prometheus.MustNewConstMetric(
		c.memoryTotalBytes,
		prometheus.GaugeValue,
		stats.Total,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryFreeBytes,
		prometheus.GaugeValue,
		stats.Free,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryAvailableBytes,
		prometheus.GaugeValue,
		stats.Available,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryBuffersBytes,
		prometheus.GaugeValue,
		stats.Buffers,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryCachedBytes,
		prometheus.GaugeValue,
		stats.Cached,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsagePercent,
		prometheus.GaugeValue,
		stats.UsagePercent,
		host,
	)

	logger.Printf("Collected memory metrics for host %s", host)
}

// parseMemoryStats 解析 /proc/meminfo 输出
func parseMemoryStats(output string) *MemoryStats {
	stats := &MemoryStats{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// 移除冒号
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		// /proc/meminfo 的值单位是 kB, 转换为字节
		valueBytes := value * 1024

		switch key {
		case "MemTotal":
			stats.Total = valueBytes
		case "MemFree":
			stats.Free = valueBytes
		case "MemAvailable":
			stats.Available = valueBytes
		case "Buffers":
			stats.Buffers = valueBytes
		case "Cached":
			stats.Cached = valueBytes
		}
	}

	// 计算使用率
	if stats.Total > 0 {
		used := stats.Total - stats.Available
		stats.UsagePercent = (used / stats.Total) * 100
	}

	return stats
}

// collectDiskMetrics 收集磁盘指标
func (c *SSHCollector) collectDiskMetrics(client *sshclient.Client, host string, ch chan<- prometheus.Metric) {
	// 执行 df 命令获取磁盘使用情况
	// -B1 表示以字节为单位显示
	output, err := client.ExecuteCommand("df -B1 -x tmpfs -x devtmpfs -x squashfs 2>/dev/null")
	if err != nil {
		logger.Printf("Failed to get disk usage on %s: %v", host, err)
		return
	}

	diskStats := parseDiskStats(output)
	logger.Printf("Found %d disk partitions on %s", len(diskStats), host)

	for _, disk := range diskStats {
		ch <- prometheus.MustNewConstMetric(
			c.diskTotalBytes,
			prometheus.GaugeValue,
			disk.Total,
			host, disk.Device, disk.MountPoint,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskUsedBytes,
			prometheus.GaugeValue,
			disk.Used,
			host, disk.Device, disk.MountPoint,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskFreeBytes,
			prometheus.GaugeValue,
			disk.Free,
			host, disk.Device, disk.MountPoint,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskUsagePercent,
			prometheus.GaugeValue,
			disk.UsagePercent,
			host, disk.Device, disk.MountPoint,
		)
	}
}

// parseDiskStats 解析 df 命令输出
func parseDiskStats(output string) []DiskStats {
	var diskStats []DiskStats
	lines := strings.Split(output, "\n")

	// 跳过第一行(表头)
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// df 输出格式: Filesystem 1B-blocks Used Available Use% Mounted on
		device := fields[0]
		total, err1 := strconv.ParseFloat(fields[1], 64)
		used, err2 := strconv.ParseFloat(fields[2], 64)
		free, err3 := strconv.ParseFloat(fields[3], 64)
		usageStr := strings.TrimSuffix(fields[4], "%")
		usagePercent, err4 := strconv.ParseFloat(usageStr, 64)
		mountPoint := fields[5]

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}

		diskStats = append(diskStats, DiskStats{
			Device:       device,
			MountPoint:   mountPoint,
			Total:        total,
			Used:         used,
			Free:         free,
			UsagePercent: usagePercent,
		})
	}

	return diskStats
}
