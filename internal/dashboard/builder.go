package dashboard

import (
	"strings"
)

// AddInstanceVariable 为 dashboard 添加 instance 变量
func AddInstanceVariable(dashboard map[string]interface{}, instances []string) error {
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

	// 根据面板类型确定标签
	label := "客户端实例"
	if title, ok := dashboard["title"].(string); ok {
		if strings.Contains(title, "服务端") || strings.Contains(title, "Server") {
			label = "服务端实例"
		}
	}

	instanceVar := map[string]interface{}{
		"name":  "instance",
		"type":  "custom",
		"label": label,
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

// AddInstanceFilterToQueries 为所有查询添加 instance 过滤
func AddInstanceFilterToQueries(dashboard map[string]interface{}) error {
	panels, ok := dashboard["panels"].([]interface{})
	if !ok {
		return nil
	}

	for _, p := range panels {
		panel, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// 跳过 text 面板
		if panel["type"] == "text" {
			continue
		}

		targets, ok := panel["targets"].([]interface{})
		if !ok {
			continue
		}

		for _, t := range targets {
			target, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

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

// FixDatasource 修复 dashboard 中的数据源引用
func FixDatasource(dashboard map[string]interface{}) {
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
