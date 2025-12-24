package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateManager 管理 dashboard 模板的加载和组合
type TemplateManager struct {
	baseDir string
}

// NewTemplateManager 创建新的模板管理器
func NewTemplateManager(baseDir string) *TemplateManager {
	return &TemplateManager{
		baseDir: baseDir,
	}
}

// LoadTemplateWithPanels 加载基础模板并组合panels片段
func (tm *TemplateManager) LoadTemplateWithPanels(baseTemplatePath string, panelDirs ...string) (map[string]interface{}, error) {
	// 加载基础模板
	dashboard, err := tm.loadBaseTemplate(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("加载基础模板失败: %w", err)
	}

	// 加载并组合panels
	panels, err := tm.loadPanels(panelDirs...)
	if err != nil {
		return nil, fmt.Errorf("加载panels失败: %w", err)
	}

	// 如果加载了panels，替换原有的panels
	if len(panels) > 0 {
		dashboard["panels"] = panels
	}

	return dashboard, nil
}

// LoadBusinessTemplateWithPanels 加载业务监控模板并组合客户端panels
// 只包含客户端功能业务相关的panels
func (tm *TemplateManager) LoadBusinessTemplateWithPanels(baseTemplatePath string, clientPanelsDir string) (map[string]interface{}, error) {
	// 加载基础模板
	dashboard, err := tm.loadBaseTemplate(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("加载基础模板失败: %w", err)
	}

	var allPanels []interface{}

	// 添加客户端指标行标题
	clientRow := map[string]interface{}{
		"collapsed": false,
		"gridPos": map[string]interface{}{
			"h": 1,
			"w": 24,
			"x": 0,
			"y": 0,
		},
		"id":     20,
		"panels": []interface{}{},
		"title":  "客户端指标",
		"type":   "row",
	}
	allPanels = append(allPanels, clientRow)

	// 加载客户端panels
	clientPanels, err := tm.loadPanelsFromDir(tm.resolvePath(clientPanelsDir))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("加载客户端panels失败: %w", err)
	}
	allPanels = append(allPanels, clientPanels...)

	// 替换原有的panels
	dashboard["panels"] = allPanels

	return dashboard, nil
}

// LoadServerMonitoringTemplateWithPanels 加载服务端监控模板并组合panels
// 包含服务端监控概览和服务端业务统计两个row
func (tm *TemplateManager) LoadServerMonitoringTemplateWithPanels(baseTemplatePath string, serverPanelsDir string) (map[string]interface{}, error) {
	// 加载基础模板
	dashboard, err := tm.loadBaseTemplate(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("加载基础模板失败: %w", err)
	}

	var allPanels []interface{}

	// 添加服务端监控概览行标题
	overviewRow := map[string]interface{}{
		"collapsed": false,
		"gridPos": map[string]interface{}{
			"h": 1,
			"w": 24,
			"x": 0,
			"y": 0,
		},
		"id":     1,
		"panels": []interface{}{},
		"title":  "服务端监控概览",
		"type":   "row",
	}
	allPanels = append(allPanels, overviewRow)

	// 加载服务端监控概览panels（前几个panel）
	serverOverviewPanels, err := tm.loadServerOverviewPanels(tm.resolvePath(serverPanelsDir))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("加载服务端监控概览panels失败: %w", err)
	}
	allPanels = append(allPanels, serverOverviewPanels...)

	// 添加业务统计行标题
	businessRow := map[string]interface{}{
		"collapsed": false,
		"gridPos": map[string]interface{}{
			"h": 1,
			"w": 24,
			"x": 0,
			"y": 13,
		},
		"id":     6,
		"panels": []interface{}{},
		"title":  "业务统计",
		"type":   "row",
	}
	allPanels = append(allPanels, businessRow)

	// 加载业务统计panels
	businessPanels, err := tm.loadBusinessPanels(tm.resolvePath(serverPanelsDir))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("加载业务统计panels失败: %w", err)
	}
	allPanels = append(allPanels, businessPanels...)

	// 替换原有的panels
	dashboard["panels"] = allPanels

	return dashboard, nil
}

// loadServerOverviewPanels 加载服务端监控概览相关的panels
func (tm *TemplateManager) loadServerOverviewPanels(dir string) ([]interface{}, error) {
	// 定义服务端监控概览的panel文件名
	overviewPanelNames := []string{
		"服务端软件版本号.json",
		"服务端健康状态.json",
		"POP端与服务端通信状态.json",
		"服务端到POP延迟.json",
	}

	return tm.loadSpecificPanels(dir, overviewPanelNames)
}

// loadBusinessPanels 加载业务统计相关的panels
func (tm *TemplateManager) loadBusinessPanels(dir string) ([]interface{}, error) {
	// 定义业务统计的panel文件名
	businessPanelNames := []string{
		"各用户订单个数.json",
	}

	return tm.loadSpecificPanels(dir, businessPanelNames)
}

// loadSpecificPanels 从目录加载指定的panel文件
func (tm *TemplateManager) loadSpecificPanels(dir string, panelNames []string) ([]interface{}, error) {
	var panels []interface{}

	for _, name := range panelNames {
		fullPath := filepath.Join(dir, name)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// 如果文件不存在，跳过
			continue
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("读取panel文件 %s 失败: %w", fullPath, err)
		}

		var panel interface{}
		if err := json.Unmarshal(data, &panel); err != nil {
			return nil, fmt.Errorf("解析panel文件 %s 失败: %w", fullPath, err)
		}

		panels = append(panels, panel)
	}

	return panels, nil
}

// loadBaseTemplate 加载基础模板（不包含panels或包含基础panels）
func (tm *TemplateManager) loadBaseTemplate(templatePath string) (map[string]interface{}, error) {
	fullPath := tm.resolvePath(templatePath)
	return LoadTemplate(fullPath)
}

// loadPanels 从多个目录加载panels片段并组合
func (tm *TemplateManager) loadPanels(panelDirs ...string) ([]interface{}, error) {
	var allPanels []interface{}

	for _, dir := range panelDirs {
		if dir == "" {
			continue
		}

		fullDir := tm.resolvePath(dir)
		panels, err := tm.loadPanelsFromDir(fullDir)
		if err != nil {
			// 如果目录不存在，跳过
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		allPanels = append(allPanels, panels...)
	}

	return allPanels, nil
}

// loadPanelsFromDir 从目录加载所有panel文件
func (tm *TemplateManager) loadPanelsFromDir(dir string) ([]interface{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var panels []interface{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只加载JSON文件
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		panel, err := tm.loadPanelFromFile(filePath)
		if err != nil {
			// 单个文件加载失败不影响其他文件
			continue
		}

		// 支持单个panel或panel数组
		if panelArray, ok := panel.([]interface{}); ok {
			panels = append(panels, panelArray...)
		} else {
			panels = append(panels, panel)
		}
	}

	return panels, nil
}

// loadPanelFromFile 从文件加载单个或一组panel
func (tm *TemplateManager) loadPanelFromFile(filePath string) (interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var panel interface{}
	if err := json.Unmarshal(data, &panel); err != nil {
		return nil, err
	}

	return panel, nil
}

// resolvePath 解析路径（相对路径转绝对路径）
func (tm *TemplateManager) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if tm.baseDir != "" {
		return filepath.Join(tm.baseDir, path)
	}

	return path
}

// SplitTemplate 将完整的模板拆分成基础模板和panels片段
// 用于初始化拆分模板文件
func (tm *TemplateManager) SplitTemplate(templatePath string, outputBase string, panelsDir string) error {
	// 加载完整模板
	dashboard, err := tm.loadBaseTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("加载模板失败: %w", err)
	}

	// 提取panels
	panels, ok := dashboard["panels"].([]interface{})
	if !ok {
		return fmt.Errorf("模板中没有panels字段")
	}

	// 创建panels目录
	if err := os.MkdirAll(panelsDir, 0755); err != nil {
		return fmt.Errorf("创建panels目录失败: %w", err)
	}

	// 保存每个panel到单独文件
	for i, panel := range panels {
		panelMap, ok := panel.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取panel标题作为文件名
		title := fmt.Sprintf("panel_%d", i+1)
		if panelTitle, ok := panelMap["title"].(string); ok && panelTitle != "" {
			// 清理标题作为文件名
			title = sanitizeFileName(panelTitle)
		}

		panelFile := filepath.Join(panelsDir, fmt.Sprintf("%s.json", title))
		panelData, err := json.MarshalIndent(panelMap, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化panel失败: %w", err)
		}

		if err := os.WriteFile(panelFile, panelData, 0644); err != nil {
			return fmt.Errorf("保存panel文件失败: %w", err)
		}
	}

	// 移除panels，保存基础模板
	delete(dashboard, "panels")
	baseData, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化基础模板失败: %w", err)
	}

	if err := os.WriteFile(outputBase, baseData, 0644); err != nil {
		return fmt.Errorf("保存基础模板失败: %w", err)
	}

	return nil
}

// sanitizeFileName 清理文件名，移除不合法字符
func sanitizeFileName(name string) string {
	// 替换空格和特殊字符
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	// 限制长度
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}
