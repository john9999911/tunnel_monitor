package cmd

import (
	"github.com/spf13/cobra"
	"tunnel-monitor/internal/service"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止监控服务",
	Long:  "停止 Prometheus 和 Grafana 服务",
	RunE: func(cmd *cobra.Command, args []string) error {
		return service.StopAll()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
