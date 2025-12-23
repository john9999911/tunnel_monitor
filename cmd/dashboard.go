package cmd

import (
	"tunnel-monitor/internal/dashboard"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "管理监控面板",
	Long:  "创建和管理 Grafana 监控面板",
}

var createBusinessCmd = &cobra.Command{
	Use:   "create",
	Short: "创建IPTunnel业务监控面板",
	Long:  "创建业务监控面板，包含客户端和服务端所有指标，支持按带宽线路筛选",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.CreateBusinessDashboard()
	},
}

var createAllCmd = &cobra.Command{
	Use:   "create-all",
	Short: "创建所有监控面板",
	Long:  "创建业务监控面板（统一包含客户端、服务端和数据库指标）",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.CreateAllDashboards()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有面板",
	Long:  "列出 Grafana 中的所有监控面板",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.ListDashboards()
	},
}

func init() {
	// 主要命令
	dashboardCmd.AddCommand(createBusinessCmd)
	dashboardCmd.AddCommand(createAllCmd)
	dashboardCmd.AddCommand(listCmd)
	
	rootCmd.AddCommand(dashboardCmd)
}

