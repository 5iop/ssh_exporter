package main

import (
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
	listenAddr = flag.String("listen", ":9109", "Address to listen on for HTTP requests")
)

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

	// 设置HTTP处理器
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	// 启动HTTP服务器
	logger.Printf("Starting HTTP server on %s", *listenAddr)
	logger.Printf("Metrics available at: http://%s/metrics", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		logger.Fatalf("Failed to start HTTP server: %v", err)
	}
}
