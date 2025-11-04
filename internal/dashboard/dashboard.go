package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"tunnel-monitor/internal/config"
)

func CreateUnifiedDashboard() error {
	fmt.Println("📊 创建统一客户端监控面板...")
	
	cfg := config.Global
	
	// 读取客户端模板
	templateFile := cfg.Dashboards.ClientTemplate
	if templateFile == "" {
		templateFile = "./dashboards/client-template.json"
	}
	
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %w", err)
	}
	
	var dashboard map[string]interface{}
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return fmt.Errorf("解析模板文件失败: %w", err)
	}
	
	// 获取所有客户端实例
	instances, err := getClientInstances()
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
	if err := addInstanceVariable(dashboard, instances); err != nil {
		return fmt.Errorf("添加实例变量失败: %w", err)
	}
	
	// 为所有查询添加 instance 过滤
	if err := addInstanceFilterToQueries(dashboard); err != nil {
		return fmt.Errorf("添加实例过滤失败: %w", err)
	}
	
	// 修复数据源引用
	fixDatasource(dashboard)
	
	// 导入到 Grafana
	if err := importDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}
	
	fmt.Println("✅ 统一客户端监控面板创建成功")
	return nil
}

func CreateServerDashboard() error {
	fmt.Println("📊 创建服务端监控面板...")
	
	cfg := config.Global
	
	// 读取服务端模板
	templateFile := cfg.Dashboards.ServerTemplate
	if templateFile == "" {
		templateFile = "./dashboards/server-template.json"
	}
	
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %w", err)
	}
	
	var dashboard map[string]interface{}
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return fmt.Errorf("解析模板文件失败: %w", err)
	}
	
	// 设置 UID
	dashboard["uid"] = cfg.Dashboards.ServerUID
	
	// 修复数据源引用
	fixDatasource(dashboard)
	
	// 导入到 Grafana
	if err := importDashboard(dashboard); err != nil {
		return fmt.Errorf("导入面板失败: %w", err)
	}
	
	fmt.Println("✅ 服务端监控面板创建成功")
	return nil
}

func ListDashboards() error {
	fmt.Println("📊 监控面板列表:")
	fmt.Println()
	
	dashboards, err := getDashboards()
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

func getClientInstances() ([]string, error) {
	// 从 Prometheus 查询客户端实例
	promURL := config.Global.Prometheus.URL
	query := `wg_interface_up{job=~"tunnel-client.*"}`
	
	encodedQuery := strings.ReplaceAll(query, " ", "%20")
	url := fmt.Sprintf("%s/api/v1/query?query=%s", promURL, encodedQuery)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	
	status := getString(result, "status")
	if status != "success" {
		return nil, fmt.Errorf("查询失败: %v", result)
	}
	
	dataObj := result["data"].(map[string]interface{})
	results := dataObj["result"].([]interface{})
	
	instances := make(map[string]bool)
	for _, r := range results {
		metric := r.(map[string]interface{})["metric"].(map[string]interface{})
		instance := getString(metric, "instance")
		if instance != "" && strings.Contains(instance, ":") {
			instances[instance] = true
		}
	}
	
	resultList := make([]string, 0, len(instances))
	for inst := range instances {
		resultList = append(resultList, inst)
	}
	
	return resultList, nil
}

func addInstanceVariable(dashboard map[string]interface{}, instances []string) error {
	templating := make(map[string]interface{})
	if val, ok := dashboard["templating"]; ok {
		templating = val.(map[string]interface{})
	}
	
	list := []interface{}{}
	if val, ok := templating["list"]; ok {
		list = val.([]interface{})
	}
	
	// 创建选项
	options := []map[string]interface{}{}
	for i, inst := range instances {
		options = append(options, map[string]interface{}{
			"text":     inst,
			"value":    inst,
			"selected": i == 0,
		})
	}
	
	instanceVar := map[string]interface{}{
		"name":    "instance",
		"type":    "custom",
		"label":   "客户端实例",
		"current": map[string]interface{}{
			"text":  options[0]["text"],
			"value": options[0]["value"],
		},
		"options":     options,
		"query":       strings.Join(instances, ","),
		"hide":        0,
		"includeAll":  false,
		"multi":       false,
		"refresh":     1,
		"regex":       "",
		"skipUrlSync": false,
		"sort":        0,
	}
	
	// 移除已存在的 instance 变量
	newList := []interface{}{}
	for _, v := range list {
		if varMap, ok := v.(map[string]interface{}); ok {
			if varMap["name"] != "instance" {
				newList = append(newList, v)
			}
		}
	}
	
	// 添加到开头
	newList = append([]interface{}{instanceVar}, newList...)
	
	templating["list"] = newList
	dashboard["templating"] = templating
	
	return nil
}

func addInstanceFilterToQueries(dashboard map[string]interface{}) error {
	panels := dashboard["panels"].([]interface{})
	
	for _, p := range panels {
		panel := p.(map[string]interface{})
		
		// 跳过 text 面板
		if panel["type"] == "text" {
			continue
		}
		
		targets := panel["targets"].([]interface{})
		for _, t := range targets {
			target := t.(map[string]interface{})
			
			expr, ok := target["expr"].(string)
			if !ok {
				continue
			}
			
			// 如果已经有 instance 变量，跳过
			if strings.Contains(expr, "$instance") {
				continue
			}
			
			// 移除硬编码的 instance
			expr = removeInstanceFilter(expr)
			
			// 添加 instance 变量
			expr = addInstanceVariableToQuery(expr)
			
			target["expr"] = expr
		}
	}
	
	return nil
}

func removeInstanceFilter(expr string) string {
	// 移除 instance="xxx" 或 instance='xxx'
	expr = strings.ReplaceAll(expr, `instance="[^"]*"\s*,?\s*`, "")
	expr = strings.ReplaceAll(expr, `instance='[^']*'\s*,?\s*`, "")
	expr = strings.ReplaceAll(expr, `,\s*,`, ",")
	expr = strings.ReplaceAll(expr, `{\s*,`, "{")
	expr = strings.ReplaceAll(expr, `,\s*}`, "}")
	return expr
}

func addInstanceVariableToQuery(expr string) string {
	if strings.Contains(expr, "$instance") {
		return expr
	}
	
	if strings.Contains(expr, "{") {
		idx := strings.LastIndex(expr, "{")
		after := expr[idx+1:]
		
		if strings.TrimSpace(after) == "}" {
			// 空的标签选择器
			return expr[:idx+1] + `instance="$instance"` + after
		}
		
		if !strings.HasPrefix(strings.TrimSpace(after), "}") {
			// 有内容，添加 instance 变量
			if strings.HasPrefix(strings.TrimSpace(after), ",") {
				return expr[:idx+1] + `instance="$instance"` + after
			}
			return expr[:idx+1] + `instance="$instance",` + after
		}
	}
	
	// 没有标签选择器，添加一个
	return expr + `{instance="$instance"}`
}

func fixDatasource(dashboard map[string]interface{}) {
	fixDatasourceRecursive(dashboard)
}

func fixDatasourceRecursive(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		if ds, ok := v["datasource"].(map[string]interface{}); ok {
			if uid, ok := ds["uid"].(string); ok && uid == "prometheus" {
				ds["uid"] = "ef32in03bdb0gb"
			}
		}
		for _, val := range v {
			fixDatasourceRecursive(val)
		}
	case []interface{}:
		for _, item := range v {
			fixDatasourceRecursive(item)
		}
	}
}

func importDashboard(dashboard map[string]interface{}) error {
	cfg := config.Global
	
	payload := map[string]interface{}{
		"dashboard": dashboard,
		"overwrite": true,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("%s/api/dashboards/db", cfg.Grafana.URL)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.Grafana.Username, cfg.Grafana.Password)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("导入失败: %s - %s", resp.Status, string(body))
	}
	
	return nil
}

func getDashboards() ([]map[string]interface{}, error) {
	cfg := config.Global
	
	url := fmt.Sprintf("%s/api/search?type=dash-db", cfg.Grafana.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(cfg.Grafana.Username, cfg.Grafana.Password)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var dashboards []map[string]interface{}
	if err := json.Unmarshal(data, &dashboards); err != nil {
		return nil, err
	}
	
	return dashboards, nil
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

