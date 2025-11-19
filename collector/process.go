package collector

import (
	"strings"

	"ssh_exporter/config"
	sshclient "ssh_exporter/ssh"

	"github.com/prometheus/client_golang/prometheus"
)

// collectProcessMetrics 收集进程监控指标
func (c *SSHCollector) collectProcessMetrics(client *sshclient.Client, host string, monitor config.ProcessMonitor, ch chan<- prometheus.Metric) {
	// 执行find命令获取进程cmdline（路径模式固定为 /proc/[0-9]*/cmdline）
	command := "find /proc -maxdepth 2 -name 'cmdline' -path '/proc/[0-9]*/cmdline' -exec cat {} \\; 2>/dev/null"
	output, err := client.ExecuteCommand(command)
	if err != nil {
		logger.Printf("Failed to get process cmdlines on %s: %v", host, err)
		return
	}

	// 解析输出
	cmdlines := parseProcessOutput(output)
	logger.Printf("Found %d process cmdlines on %s", len(cmdlines), host)

	// 统计每个pattern的出现次数
	for _, pattern := range monitor.Patterns {
		count := 0
		for _, cmdline := range cmdlines {
			if strings.Contains(cmdline, pattern) {
				count++
			}
		}
		logger.Printf("Host %s: pattern '%s' found %d times", host, pattern, count)

		// 使用统一的指标描述符
		ch <- prometheus.MustNewConstMetric(
			c.processPatternCount,
			prometheus.GaugeValue,
			float64(count),
			host,
			pattern,
		)
	}
}

// parseProcessOutput 解析进程输出
func parseProcessOutput(output string) []string {
	var cmdlines []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// 将null字符替换为空格
		cleaned := strings.ReplaceAll(line, "\x00", " ")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			cmdlines = append(cmdlines, cleaned)
		}
	}

	return cmdlines
}
