package collector

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ssh_exporter/config"
	sshclient "ssh_exporter/ssh"

	"github.com/prometheus/client_golang/prometheus"
)

// FileInfo 文件信息
type FileInfo struct {
	Path         string
	Size         int64
	LastModified float64
}

// collectFileMetrics 收集文件监控指标
func (c *SSHCollector) collectFileMetrics(client *sshclient.Client, host string, monitor config.FileMonitor, ch chan<- prometheus.Metric, currentTime float64) {
	// 执行ls命令获取文件信息
	command := fmt.Sprintf("ls -lgb --full-time %s 2>/dev/null | awk '{print $4\"\\t\"$5\" \"$6\"\\t\"$8}'", monitor.Path)
	output, err := client.ExecuteCommand(command)
	if err != nil {
		logger.Printf("Failed to get file info on %s: %v", host, err)
		return
	}

	// 解析输出
	fileInfos := parseFileOutput(output, monitor.Path)
	logger.Printf("Found %d files on %s in path %s", len(fileInfos), host, monitor.Path)

	for _, info := range fileInfos {
		filename := filepath.Base(info.Path)

		// 应用标签匹配规则（如果有）
		// 注意：这里不需要添加额外的标签到metric中，因为Prometheus的Desc已经定义了固定的标签

		// 文件大小
		ch <- prometheus.MustNewConstMetric(
			c.fileSize,
			prometheus.GaugeValue,
			float64(info.Size),
			host, monitor.Path, filename,
		)

		// 最后修改时间
		ch <- prometheus.MustNewConstMetric(
			c.fileLastModified,
			prometheus.GaugeValue,
			info.LastModified,
			host, monitor.Path, filename,
		)

		// 文件年龄（分钟）
		ageMinutes := (currentTime - info.LastModified) / 60
		ch <- prometheus.MustNewConstMetric(
			c.fileAgeMinutes,
			prometheus.GaugeValue,
			ageMinutes,
			host, monitor.Path, filename,
		)
	}
}

// parseFileOutput 解析文件输出
func parseFileOutput(output string, basePath string) []FileInfo {
	var fileInfos []FileInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		// 解析文件大小
		size, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}

		// 解析时间戳 (格式: 2024-01-01 12:00:00.000000000)
		timeStr := strings.TrimSpace(parts[1])
		// 只取前26个字符 (包含到微秒)
		if len(timeStr) > 26 {
			timeStr = timeStr[:26]
		}

		// 解析时间
		t, err := time.Parse("2006-01-02 15:04:05.000000", timeStr)
		if err != nil {
			logger.Printf("Failed to parse time '%s': %v", timeStr, err)
			continue
		}

		// 文件名
		filename := strings.TrimSpace(parts[2])
		if filename == "" {
			continue
		}

		fileInfos = append(fileInfos, FileInfo{
			Path:         filename,
			Size:         size,
			LastModified: float64(t.Unix()),
		})
	}

	return fileInfos
}

// matchesPattern 检查文件名是否匹配正则表达式
func matchesPattern(filename, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Printf("Invalid regex pattern '%s': %v", pattern, err)
		return false
	}
	return re.MatchString(filename)
}
