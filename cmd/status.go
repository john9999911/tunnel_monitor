package cmd

import (
	"github.com/spf13/cobra"
	"tunnel-monitor/internal/service"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看监控服务状态",
	Long:  "查看 Prometheus 和 Grafana 的运行状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		return service.ShowStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

