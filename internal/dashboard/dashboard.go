package dashboard

import (
	"fmt"

	"tunnel-monitor/internal/config"
)

// CreateBusinessDashboard 创建IPTunnel业务监控面板（包含客户端和服务端指标）
func CreateBusinessDashboard() error {
	fmt.Println("📊 创建IPTunnel业务监控面板...")

	cfg := config.Global

	// 加载业务模板
	dashboard, err := LoadBusinessTemplate()
	if err != nil {
		return fmt.Errorf("加载业务模板失败: %w", err)
	}

	// 设置面板标题和UID
	dashboard["title"] = "IPTunnel 业务监控"
	dashboard["uid"] = cfg.Dashboards.BusinessUID
	if dashboard["uid"] == "" {
		dashboard["uid"] = "iptunnel-business"
	}

	// 修复数据源引用
	FixDatasource(dashboard)

	// 导入到 Grafana
	if err := ImportDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}

	fmt.Println("✅ IPTunnel业务监控面板创建成功")
	fmt.Println("💡 提示：")
	fmt.Println("   - 使用'带宽线路'下拉框筛选特定线路")
	fmt.Println("   - 选择'All'显示所有线路数据")
	fmt.Println("   - 客户端数据由服务端转发，通过exported_instance标签区分")
	return nil
}

// ListDashboards 列出所有监控面板
func ListDashboards() error {
	fmt.Println("📊 监控面板列表:")
	fmt.Println()

	dashboards, err := GetDashboards()
	if err != nil {
		return fmt.Errorf("获取面板列表失败: %w", err)
	}

	if len(dashboards) == 0 {
		fmt.Println("未找到任何面板")
		return nil
	}

	for _, db := range dashboards {
		title := getString(db, "title")
		uid := getString(db, "uid")
		url := getString(db, "url")

		fmt.Printf("📋 %s\n", title)
		fmt.Printf("   UID: %s\n", uid)
		if url != "" {
			fmt.Printf("   访问: %s%s\n", config.Global.Grafana.URL, url)
		}
		fmt.Println()
	}

	return nil
}

// CreateAllDashboards 创建所有监控面板（仅创建业务监控面板）
func CreateAllDashboards() error {
	fmt.Println("🚀 开始创建监控面板...")
	fmt.Println()

	// 只创建业务监控面板
	if err := CreateBusinessDashboard(); err != nil {
		return fmt.Errorf("业务面板创建失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 监控面板创建完成！")
	return nil
}

