package main

import (
	"crypto/subtle"
	"flag"
	"log"
	"net/http"
	"os"

	"ssh_exporter/collector"
	"ssh_exporter/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configFile = flag.String("config", "run/config.yaml", "Path to configuration file")
	listenAddr = flag.String("listen", "", "Address to listen on for HTTP requests (overrides config file)")
)

// basicAuth 提供HTTP基本认证中间件
func basicAuth(username, password string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="SSH Exporter"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	// 设置日志
	logger := log.New(os.Stdout, "[Main] ", log.LstdFlags)

	// 加载配置
	logger.Printf("Loading configuration from %s", *configFile)
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}
	logger.Printf("Loaded configuration for %d hosts", len(cfg.Hosts))

	// 创建并注册collector
	sshCollector := collector.NewSSHCollector(cfg)
	prometheus.MustRegister(sshCollector)
	logger.Println("SSH Collector registered")

	// 确定监听地址 - 命令行参数优先于配置文件
	listen := *listenAddr
	if listen == "" {
		if cfg.Listen != "" {
			listen = cfg.Listen
		} else {
			listen = ":9109" // 默认端口
		}
	}

	// 设置HTTP处理器
	metricsHandler := promhttp.Handler()
	indexHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
<head><title>SSH Exporter</title></head>
<body>
<h1>SSH Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<h2>Configuration</h2>
<p>Monitoring ` + string(rune(len(cfg.Hosts))) + ` hosts</p>
</body>
</html>`))
	})

	// 应用HTTP基本认证（如果配置了的话）
	var finalMetricsHandler http.Handler = metricsHandler
	var finalIndexHandler http.Handler = indexHandler
	if cfg.HTTPAuth != nil && cfg.HTTPAuth.Username != "" && cfg.HTTPAuth.Password != "" {
		logger.Println("HTTP basic authentication enabled")
		finalMetricsHandler = basicAuth(cfg.HTTPAuth.Username, cfg.HTTPAuth.Password, metricsHandler)
		finalIndexHandler = basicAuth(cfg.HTTPAuth.Username, cfg.HTTPAuth.Password, indexHandler)
	}

	http.Handle("/metrics", finalMetricsHandler)
	http.Handle("/", finalIndexHandler)

	// 启动HTTP服务器
	logger.Printf("Starting HTTP server on %s", listen)
	logger.Printf("Metrics available at: http://%s/metrics", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		logger.Fatalf("Failed to start HTTP server: %v", err)
	}
}
