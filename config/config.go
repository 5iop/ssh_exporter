package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 总配置结构
type Config struct {
	Hosts []HostConfig `yaml:"hosts"`
}

// HostConfig 主机配置
type HostConfig struct {
	Host     string          `yaml:"host"`
	User     string          `yaml:"user"`
	Password string          `yaml:"password"`
	Port     int             `yaml:"port"` // SSH端口，默认22
	Monitors MonitorConfig   `yaml:"monitors"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Processes []ProcessMonitor `yaml:"processes"`
	Files     []FileMonitor    `yaml:"files"`
	Stat      bool             `yaml:"stat"` // 系统统计监控(CPU、内存、磁盘)
}

// ProcessMonitor 进程监控配置
type ProcessMonitor struct {
	PathPattern string   `yaml:"path_pattern"` // 例如：/proc/[0-9]*/cmdline
	Patterns    []string `yaml:"patterns"`     // 要搜索的模式列表
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
