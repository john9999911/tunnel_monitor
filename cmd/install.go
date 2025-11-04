package cmd

import (
	"tunnel-monitor/internal/installer"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装 Prometheus 和 Grafana",
	Long:  "自动检测操作系统并安装 Prometheus 和 Grafana",
	RunE: func(cmd *cobra.Command, args []string) error {
		return installer.InstallAll()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
