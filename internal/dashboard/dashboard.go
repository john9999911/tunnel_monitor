package dashboard

import (
	"fmt"

	"tunnel-monitor/internal/config"
)

// CreateUnifiedDashboard 创建统一客户端监控面板
func CreateUnifiedDashboard() error {
	fmt.Println("📊 创建统一客户端监控面板...")

	cfg := config.Global

	// 加载客户端模板
	dashboard, err := LoadClientTemplate()
	if err != nil {
		return fmt.Errorf("加载模板失败: %w", err)
	}

	// 获取所有客户端实例
	instances, err := GetClientInstances()
	if err != nil {
		return fmt.Errorf("获取客户端实例失败: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("未找到任何客户端实例")
	}

	fmt.Printf("✅ 发现 %d 个客户端实例\n", len(instances))

	// 修改 dashboard
	dashboard["title"] = "POP 客户端统一监控面板"
	dashboard["uid"] = cfg.Dashboards.UnifiedUID

	// 添加 instance 变量
	if err := AddInstanceVariable(dashboard, instances); err != nil {
		return fmt.Errorf("添加实例变量失败: %w", err)
	}

	// 为所有查询添加 instance 过滤
	if err := AddInstanceFilterToQueries(dashboard); err != nil {
		return fmt.Errorf("添加实例过滤失败: %w", err)
	}

	// 修复数据源引用
	FixDatasource(dashboard)

	// 导入到 Grafana
	if err := ImportDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}

	fmt.Println("✅ 统一客户端监控面板创建成功")
	return nil
}

// CreateServerDashboard 创建统一服务端监控面板
func CreateServerDashboard() error {
	fmt.Println("📊 创建统一服务端监控面板...")

	cfg := config.Global

	// 加载服务端模板
	dashboard, err := LoadServerTemplate()
	if err != nil {
		return fmt.Errorf("加载模板失败: %w", err)
	}

	// 获取所有服务端实例
	instances, err := GetServerInstances()
	if err != nil {
		return fmt.Errorf("获取服务端实例失败: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("未找到任何服务端实例")
	}

	fmt.Printf("✅ 发现 %d 个服务端实例\n", len(instances))

	// 修改 dashboard
	dashboard["title"] = "Tunnel Server 统一监控面板"
	dashboard["uid"] = cfg.Dashboards.ServerUID

	// 添加 instance 变量
	if err := AddInstanceVariable(dashboard, instances); err != nil {
		return fmt.Errorf("添加实例变量失败: %w", err)
	}

	// 为所有查询添加 instance 过滤
	if err := AddInstanceFilterToQueries(dashboard); err != nil {
		return fmt.Errorf("添加实例过滤失败: %w", err)
	}

	// 修复数据源引用
	FixDatasource(dashboard)

	// 导入到 Grafana
	if err := ImportDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}

	fmt.Println("✅ 统一服务端监控面板创建成功")
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

// CreateDatabaseDashboard 创建数据库监控面板
func CreateDatabaseDashboard() error {
	fmt.Println("📊 创建数据库监控面板...")

	cfg := config.Global

	// 加载数据库模板
	dashboard, err := LoadDatabaseTemplate()
	if err != nil {
		return fmt.Errorf("加载模板失败: %w", err)
	}

	// 设置 UID 和标题
	dashboard["uid"] = "tunnel-database"
	if cfg.Dashboards.DatabaseUID != "" {
		dashboard["uid"] = cfg.Dashboards.DatabaseUID
	}

	// 修复数据源引用
	FixDatasource(dashboard)

	// 导入到 Grafana
	if err := ImportDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}

	fmt.Println("✅ 数据库监控面板创建成功")
	return nil
}

// CreateAllDashboards 创建所有监控面板
func CreateAllDashboards() error {
	fmt.Println("🚀 开始创建所有监控面板...")
	fmt.Println()

	// 创建数据库面板
	if err := CreateDatabaseDashboard(); err != nil {
		fmt.Printf("⚠️  数据库面板创建失败: %v\n", err)
	} else {
		fmt.Println()
	}

	// 创建客户端面板
	if err := CreateUnifiedDashboard(); err != nil {
		fmt.Printf("⚠️  客户端面板创建失败: %v\n", err)
	} else {
		fmt.Println()
	}

	// 创建服务端面板
	if err := CreateServerDashboard(); err != nil {
		fmt.Printf("⚠️  服务端面板创建失败: %v\n", err)
	} else {
		fmt.Println()
	}

	fmt.Println("✅ 所有监控面板创建完成！")
	return nil
}
