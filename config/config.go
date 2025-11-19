package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 总配置结构
type Config struct {
	Listen       string       `yaml:"listen"`        // HTTP监听地址，例如 ":9100"
	MetricPrefix string       `yaml:"metric_prefix"` // 指标名称前缀（可选），例如 "ssh_exporter_"
	HTTPAuth     *HTTPAuth    `yaml:"http_auth"`     // HTTP基本认证配置（可选）
	Hosts        []HostConfig `yaml:"hosts"`
}

// HTTPAuth HTTP基本认证配置
type HTTPAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// HostConfig 主机配置
type HostConfig struct {
	Host           string        `yaml:"host"`
	User           string        `yaml:"user"`
	Password       string        `yaml:"password"`        // SSH密码（可选，如果使用私钥则不需要）
	PrivateKeyPath string        `yaml:"private_key"`     // SSH私钥路径（可选）
	Port           int           `yaml:"port"`            // SSH端口，默认22
	Monitors       MonitorConfig `yaml:"monitors"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Processes []ProcessMonitor `yaml:"processes"`
	Files     []FileMonitor    `yaml:"files"`
	Stat      bool             `yaml:"stat"` // 系统统计监控(CPU、内存、磁盘)
}

// ProcessMonitor 进程监控配置
type ProcessMonitor struct {
	Patterns []string `yaml:"patterns"` // 要搜索的进程名称模式列表
}

// FileMonitor 文件监控配置
type FileMonitor struct {
	Path   string       `yaml:"path"`   // 要监控的目录
	Labels []FileLabel  `yaml:"labels"` // 文件标签匹配规则
}

// FileLabel 文件标签配置
type FileLabel struct {
	Pattern string `yaml:"pattern"` // 正则表达式
	Name    string `yaml:"name"`    // 标签名称
	Value   string `yaml:"value"`   // 标签值
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认端口
	for i := range config.Hosts {
		if config.Hosts[i].Port == 0 {
			config.Hosts[i].Port = 22
		}
	}

	return &config, nil
}
