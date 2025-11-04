package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tunnel-monitor/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tunnel-monitor",
	Short: "Tunnel Monitor - 监控面板管理工具",
	Long: `Tunnel Monitor 是一个独立的监控面板管理工具，用于管理 Prometheus 和 Grafana。

主要功能：
- 安装和管理 Prometheus 和 Grafana
- 创建和管理监控面板
- 配置服务端和客户端监控 URL`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ./config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
	}
	
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
}

