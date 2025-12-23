package prometheus

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"tunnel-monitor/internal/config"
)

type PrometheusConfig struct {
	Global struct {
		ScrapeInterval     string `yaml:"scrape_interval"`
		EvaluationInterval string `yaml:"evaluation_interval,omitempty"`
	} `yaml:"global"`

	RuleFiles     []string       `yaml:"rule_files,omitempty"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

type ScrapeConfig struct {
	JobName        string         `yaml:"job_name"`
	StaticConfigs  []StaticConfig `yaml:"static_configs"`
	MetricsPath    string         `yaml:"metrics_path,omitempty"`
	ScrapeInterval string         `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  string         `yaml:"scrape_timeout,omitempty"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets"`
}

func UpdateConfig() error {
	fmt.Println("📝 更新 Prometheus 配置...")

	cfg := config.Global

	configFile := cfg.Prometheus.ConfigFile
	if configFile == "" {
		configFile = "./prometheus.yml"
	}

	// 读取现有配置（如果存在）
	promCfg := &PrometheusConfig{}
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			if err := yaml.Unmarshal(data, promCfg); err == nil {
				fmt.Println("✅ 已读取现有 Prometheus 配置")
			}
		}
	}

	// 如果配置为空，设置默认值
	if promCfg.Global.ScrapeInterval == "" {
		promCfg.Global.ScrapeInterval = "15s"
		promCfg.Global.EvaluationInterval = "15s"
	}

	// 确保服务端配置存在
	serverJobExists := false
	serverJobIndex := -1
	for i, scrapeCfg := range promCfg.ScrapeConfigs {
		if scrapeCfg.JobName == "tunnel-server" {
			serverJobExists = true
			serverJobIndex = i
			break
		}
	}

	if !serverJobExists {
		// 添加服务端配置
		serverCfg := ScrapeConfig{
			JobName: "tunnel-server",
			StaticConfigs: []StaticConfig{
				{
					Targets: []string{fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)},
				},
			},
			MetricsPath:    "/metrics",
			ScrapeInterval: "5s",
			ScrapeTimeout:  "5s",
		}
		promCfg.ScrapeConfigs = append([]ScrapeConfig{serverCfg}, promCfg.ScrapeConfigs...)
		fmt.Println("✅ 已添加服务端配置")
	} else {
		// 更新服务端配置
		promCfg.ScrapeConfigs[serverJobIndex].StaticConfigs[0].Targets[0] = fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)
		fmt.Println("✅ 已更新服务端配置")
	}

	// 客户端配置已废弃：客户端监控数据现在由服务端转发到Prometheus
	// 不再需要单独配置客户端targets
	fmt.Println("💡 提示：客户端监控数据由服务端统一转发，无需配置客户端列表")

	// 写入配置文件
	data, err := yaml.Marshal(promCfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✅ Prometheus 配置已更新: %s\n", configFile)
	return nil
}
