package prometheus

import (
	"fmt"
	"os"
	"strings"

	"tunnel-monitor/internal/config"
	"gopkg.in/yaml.v3"
)

type PrometheusConfig struct {
	Global struct {
		ScrapeInterval    string `yaml:"scrape_interval"`
		EvaluationInterval string `yaml:"evaluation_interval,omitempty"`
	} `yaml:"global"`
	
	RuleFiles     []string      `yaml:"rule_files,omitempty"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

type ScrapeConfig struct {
	JobName        string        `yaml:"job_name"`
	StaticConfigs  []StaticConfig `yaml:"static_configs"`
	MetricsPath    string        `yaml:"metrics_path,omitempty"`
	ScrapeInterval string        `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  string        `yaml:"scrape_timeout,omitempty"`
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
	
	// 如果配置了客户端，添加到配置中
	if len(cfg.Clients) > 0 {
		// 查找或创建客户端 job
		clientJobExists := false
		clientJobIndex := -1
		for i, scrapeCfg := range promCfg.ScrapeConfigs {
			if scrapeCfg.JobName == "tunnel-client-pop" {
				clientJobExists = true
				clientJobIndex = i
				break
			}
		}
		
		if clientJobExists {
			// 更新现有客户端 job 的 targets
			targets := []string{}
			for _, client := range cfg.Clients {
				instance := client.Instance
				if instance == "" {
					instance = strings.TrimPrefix(strings.TrimPrefix(client.MetricsURL, "http://"), "https://")
				}
				targets = append(targets, instance)
			}
			if len(promCfg.ScrapeConfigs[clientJobIndex].StaticConfigs) > 0 {
				promCfg.ScrapeConfigs[clientJobIndex].StaticConfigs[0].Targets = targets
			} else {
				promCfg.ScrapeConfigs[clientJobIndex].StaticConfigs = []StaticConfig{{Targets: targets}}
			}
			fmt.Printf("✅ 已更新客户端配置（%d 个客户端）\n", len(cfg.Clients))
		} else {
			// 创建新的客户端 job
			targets := []string{}
			for _, client := range cfg.Clients {
				instance := client.Instance
				if instance == "" {
					instance = strings.TrimPrefix(strings.TrimPrefix(client.MetricsURL, "http://"), "https://")
				}
				targets = append(targets, instance)
			}
			
			clientCfg := ScrapeConfig{
				JobName: "tunnel-client-pop",
				StaticConfigs: []StaticConfig{
					{
						Targets: targets,
					},
				},
				MetricsPath:    "/metrics",
				ScrapeInterval: "5s",
				ScrapeTimeout:  "5s",
			}
			promCfg.ScrapeConfigs = append(promCfg.ScrapeConfigs, clientCfg)
			fmt.Printf("✅ 已添加客户端配置（%d 个客户端）\n", len(cfg.Clients))
		}
	}
	
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

