package cmd

import (
	"tunnel-monitor/internal/service"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动监控服务",
	Long:  "启动 Prometheus 和 Grafana 服务",
	RunE: func(cmd *cobra.Command, args []string) error {
		return service.StartAll()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
