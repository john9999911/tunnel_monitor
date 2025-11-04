package dashboard

import (
	"encoding/json"
	"fmt"
	"os"

	"tunnel-monitor/internal/config"
)

// LoadTemplate 加载 dashboard 模板文件
func LoadTemplate(templatePath string) (map[string]interface{}, error) {
	if templatePath == "" {
		return nil, fmt.Errorf("模板路径不能为空")
	}

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("读取模板文件失败: %w", err)
	}

	var dashboard map[string]interface{}
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return nil, fmt.Errorf("解析模板文件失败: %w", err)
	}

	return dashboard, nil
}

// LoadClientTemplate 加载客户端模板
// 优先尝试使用拆分后的模板（基础模板+panels目录），如果不存在则使用完整模板
func LoadClientTemplate() (map[string]interface{}, error) {
	cfg := config.Global
	templateFile := cfg.Dashboards.ClientTemplate
	if templateFile == "" {
		templateFile = "./dashboards/client-template.json"
	}

	// 尝试使用拆分模板
	baseTemplate := templateFile
	panelsDir := "./dashboards/panels/client"
	
	// 如果存在panels目录，使用模板管理器加载
	if _, err := os.Stat(panelsDir); err == nil {
		tm := NewTemplateManager("")
		return tm.LoadTemplateWithPanels(baseTemplate, panelsDir)
	}

	// 否则使用完整模板
	return LoadTemplate(templateFile)
}

// LoadServerTemplate 加载服务端模板
// 优先尝试使用拆分后的模板（基础模板+panels目录），如果不存在则使用完整模板
func LoadServerTemplate() (map[string]interface{}, error) {
	cfg := config.Global
	templateFile := cfg.Dashboards.ServerTemplate
	if templateFile == "" {
		templateFile = "./dashboards/server-template.json"
	}

	// 尝试使用拆分模板
	baseTemplate := templateFile
	panelsDir := "./dashboards/panels/server"
	
	// 如果存在panels目录，使用模板管理器加载
	if _, err := os.Stat(panelsDir); err == nil {
		tm := NewTemplateManager("")
		return tm.LoadTemplateWithPanels(baseTemplate, panelsDir)
	}

	// 否则使用完整模板
	return LoadTemplate(templateFile)
}

// LoadDatabaseTemplate 加载数据库模板
// 优先尝试使用拆分后的模板（基础模板+panels目录），如果不存在则使用完整模板
func LoadDatabaseTemplate() (map[string]interface{}, error) {
	cfg := config.Global
	templateFile := cfg.Dashboards.DatabaseTemplate
	if templateFile == "" {
		templateFile = "./dashboards/database-template.json"
	}

	// 尝试使用拆分模板
	baseTemplate := templateFile
	panelsDir := "./dashboards/panels/database"
	
	// 如果存在panels目录，使用模板管理器加载
	if _, err := os.Stat(panelsDir); err == nil {
		tm := NewTemplateManager("")
		return tm.LoadTemplateWithPanels(baseTemplate, panelsDir)
	}

	// 否则使用完整模板
	return LoadTemplate(templateFile)
}
