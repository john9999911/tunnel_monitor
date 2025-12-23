package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Global *Config

type Config struct {
	Prometheus struct {
		URL        string `yaml:"url"`
		Port       int    `yaml:"port"`
		DataDir    string `yaml:"data_dir"`
		ConfigFile string `yaml:"config_file"`
	} `yaml:"prometheus"`

	Grafana struct {
		URL           string `yaml:"url"`
		Port          int    `yaml:"port"`
		Username      string `yaml:"username"`
		Password      string `yaml:"password"`
		APIKey        string `yaml:"api_key"`
		PrometheusUID string `yaml:"prometheus_uid"` // Grafana中Prometheus数据源UID
	} `yaml:"grafana"`

	Server struct {
		MetricsURL string `yaml:"metrics_url"`
		Port       int    `yaml:"port"`
	} `yaml:"server"`

	// MySQL数据源配置
	MySQL struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		UID      string `yaml:"uid"` // Grafana数据源UID
	} `yaml:"mysql"`

	Dashboards struct {
		ServerTemplate   string `yaml:"server_template"`
		ClientTemplate   string `yaml:"client_template"`
		DatabaseTemplate string `yaml:"database_template"`
		BusinessTemplate string `yaml:"business_template"`
		UnifiedUID       string `yaml:"unified_uid"`
		ServerUID        string `yaml:"server_uid"`
		DatabaseUID      string `yaml:"database_uid"`
		BusinessUID      string `yaml:"business_uid"`
	} `yaml:"dashboards"`
}

var configFile = "./config.yaml"

func SetConfigFile(path string) {
	configFile = path
}

func Load() error {
	Global = &Config{}

	// 设置默认值
	setDefaults()

	// 如果配置文件存在，读取它
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}

		if err := yaml.Unmarshal(data, Global); err != nil {
			return fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	return nil
}

func setDefaults() {
	Global.Prometheus.URL = "http://localhost:9090"
	Global.Prometheus.Port = 9090
	Global.Prometheus.DataDir = "./prometheus_data"
	Global.Prometheus.ConfigFile = "./config/monitoring/prometheus.yml"

	Global.Grafana.URL = "http://localhost:3000"
	Global.Grafana.Port = 3000
	Global.Grafana.Username = "admin"
	Global.Grafana.Password = "admin"
	Global.Grafana.PrometheusUID = "prometheus-datasource" // 默认Prometheus数据源UID

	// MySQL默认值
	Global.MySQL.Host = "localhost"
	Global.MySQL.Port = 3306
	Global.MySQL.Database = "iptunnel"
	Global.MySQL.Username = "root"
	Global.MySQL.Password = ""
	Global.MySQL.UID = "mysql-datasource" // 默认MySQL数据源UID

	Global.Server.MetricsURL = "http://localhost:8001/metrics"
	Global.Server.Port = 8001

	Global.Dashboards.ServerTemplate = "./dashboards/server-template.json"
	Global.Dashboards.ClientTemplate = "./dashboards/client-template.json"
	Global.Dashboards.DatabaseTemplate = "./dashboards/database-template.json"
	Global.Dashboards.UnifiedUID = "pop-clients-unified"
	Global.Dashboards.ServerUID = "tunnel-server"
	Global.Dashboards.DatabaseUID = "tunnel-database"
}

func Save() error {
	data, err := yaml.Marshal(Global)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
