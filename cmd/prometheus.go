package cmd

import (
	"tunnel-monitor/internal/prometheus"

	"github.com/spf13/cobra"
)

var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "管理 Prometheus 配置",
	Long:  "更新和管理 Prometheus 配置",
}

var updateConfigCmd = &cobra.Command{
	Use:   "update-config",
	Short: "更新 Prometheus 配置",
	Long:  "根据配置更新 Prometheus 配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		return prometheus.UpdateConfig()
	},
}

func init() {
	prometheusCmd.AddCommand(updateConfigCmd)
	rootCmd.AddCommand(prometheusCmd)
}
