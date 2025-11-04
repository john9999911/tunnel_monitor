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

var createUnifiedCmd = &cobra.Command{
	Use:   "create-unified",
	Short: "创建统一客户端监控面板",
	Long:  "创建一个统一的面板，使用变量选择不同的客户端实例",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.CreateUnifiedDashboard()
	},
}

var createServerCmd = &cobra.Command{
	Use:   "create-server",
	Short: "创建统一服务端监控面板",
	Long:  "创建统一的服务端监控面板，支持多服务端部署（使用变量选择实例）",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.CreateServerDashboard()
	},
}

var createDatabaseCmd = &cobra.Command{
	Use:   "create-database",
	Short: "创建数据库监控面板",
	Long:  "创建数据库监控面板",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.CreateDatabaseDashboard()
	},
}

var createAllCmd = &cobra.Command{
	Use:   "create-all",
	Short: "创建所有监控面板",
	Long:  "一次性创建客户端、服务端和数据库监控面板",
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
	dashboardCmd.AddCommand(createUnifiedCmd)
	dashboardCmd.AddCommand(createServerCmd)
	dashboardCmd.AddCommand(createDatabaseCmd)
	dashboardCmd.AddCommand(createAllCmd)
	dashboardCmd.AddCommand(listCmd)
	rootCmd.AddCommand(dashboardCmd)
}
