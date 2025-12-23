package dashboard

import (
	"strings"

	"tunnel-monitor/internal/config"
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
			if uid, ok := ds["uid"].(string); ok {
				switch uid {
				case "{{PROMETHEUS_UID}}":
					// 替换Prometheus数据源UID占位符
					ds["uid"] = config.Global.Grafana.PrometheusUID
				case "{{MYSQL_UID}}":
					// 替换MySQL数据源UID占位符
					ds["uid"] = config.Global.MySQL.UID
				}
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

	// 先处理有 { 的情况，在标签选择器中添加instance
	braceIdx := strings.Index(expr, "{")
	if braceIdx > 0 {
		afterBrace := expr[braceIdx+1:]
		closeBrace := strings.Index(afterBrace, "}")
		if closeBrace >= 0 {
			labels := strings.TrimSpace(afterBrace[:closeBrace])
			if labels == "" {
				return expr[:braceIdx+1] + `instance="$instance"` + expr[braceIdx+1:]
			} else if !strings.Contains(labels, "instance=") {
				return expr[:braceIdx+1] + `instance="$instance",` + expr[braceIdx+1:]
			}
		}
		return expr // 已经有instance，返回
	}

	// 处理聚合函数（如 sum(metric), count(metric) 等）
	// 常见的聚合函数
	aggFuncs := []string{"sum", "count", "avg", "max", "min", "stddev", "stdvar", "topk", "bottomk", "quantile"}
	for _, funcName := range aggFuncs {
		prefix := funcName + "("
		if strings.HasPrefix(expr, prefix) {
			// 找到第一个 ( 的位置
			openParen := len(funcName)
			// 查找对应的 ) 的位置
			closeParen := findMatchingCloseParen(expr, openParen)
			if closeParen > openParen {
				// 提取函数内部的表达式
				innerExpr := expr[openParen+1 : closeParen]
				// 检查是否有 by/without 子句
				byIdx := strings.Index(innerExpr, " by ")
				withoutIdx := strings.Index(innerExpr, " without ")
				if byIdx > 0 || withoutIdx > 0 {
					// 有 by/without，在指标名称上添加instance
					sepIdx := byIdx
					if withoutIdx > 0 && (byIdx < 0 || withoutIdx < byIdx) {
						sepIdx = withoutIdx
					}
					metricExpr := strings.TrimSpace(innerExpr[:sepIdx])
					byClause := innerExpr[sepIdx:]
					// 在metricExpr上添加instance
					metricExpr = addInstanceToMetricExpr(metricExpr)
					return expr[:openParen+1] + metricExpr + byClause + expr[closeParen:]
				} else {
					// 没有 by/without，直接在指标表达式上添加instance
					innerExpr = addInstanceToMetricExpr(innerExpr)
					return expr[:openParen+1] + innerExpr + expr[closeParen:]
				}
			}
		}
	}

	// 处理没有 { 但有 [ 的情况（如 rate(metric[5m])）
	bracketIdx := strings.Index(expr, "[")
	if bracketIdx > 0 {
		// 在 [ 之前查找指标名称
		metricEnd := findMetricNameEnd(expr, bracketIdx-1)
		if metricEnd < bracketIdx {
			// 在 [ 之前插入 {instance="$instance"}
			return expr[:bracketIdx] + `{instance="$instance"}` + expr[bracketIdx:]
		}
	}

	// 处理简单的指标名称（没有 { 和 [）
	metricEnd := findMetricNameEnd(expr, len(expr)-1)
	if metricEnd == len(expr) {
		// 整个表达式都是指标名称
		return expr + `{instance="$instance"}`
	} else if metricEnd > 0 && metricEnd < len(expr) {
		// 在指标名称后添加标签选择器
		return expr[:metricEnd] + `{instance="$instance"}` + expr[metricEnd:]
	}

	// 如果无法处理，返回原表达式
	return expr
}

// findMatchingCloseParen 找到匹配的右括号
func findMatchingCloseParen(expr string, openParen int) int {
	depth := 1
	for i := openParen + 1; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// addInstanceToMetricExpr 在指标表达式上添加instance过滤
func addInstanceToMetricExpr(metricExpr string) string {
	metricExpr = strings.TrimSpace(metricExpr)

	// 如果已经有 {，在里面添加instance
	braceIdx := strings.Index(metricExpr, "{")
	if braceIdx > 0 {
		afterBrace := metricExpr[braceIdx+1:]
		closeBrace := strings.Index(afterBrace, "}")
		if closeBrace >= 0 {
			labels := strings.TrimSpace(afterBrace[:closeBrace])
			if labels == "" {
				return metricExpr[:braceIdx+1] + `instance="$instance"` + metricExpr[braceIdx+1:]
			} else if !strings.Contains(labels, "instance=") {
				return metricExpr[:braceIdx+1] + `instance="$instance",` + metricExpr[braceIdx+1:]
			}
		}
		return metricExpr
	}

	// 如果有 [，在 [ 之前添加 {instance="$instance"}
	bracketIdx := strings.Index(metricExpr, "[")
	if bracketIdx > 0 {
		metricEnd := findMetricNameEnd(metricExpr, bracketIdx-1)
		if metricEnd < bracketIdx {
			return metricExpr[:bracketIdx] + `{instance="$instance"}` + metricExpr[bracketIdx:]
		}
	}

	// 简单指标名称
	metricEnd := findMetricNameEnd(metricExpr, len(metricExpr)-1)
	if metricEnd == len(metricExpr) {
		return metricExpr + `{instance="$instance"}`
	}

	return metricExpr
}

// findMetricNameEnd 反向查找指标名称的结束位置
// 返回值：指标名称结束的位置（不包含），如果整个字符串都是指标名称，返回字符串长度
func findMetricNameEnd(expr string, start int) int {
	for i := start; i >= 0; i-- {
		c := expr[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == ':' {
			continue
		}
		// 遇到非标识符字符，指标名称结束位置是下一个字符
		return i + 1
	}
	// 如果循环完成，说明从start到0都是指标名称的一部分
	// 返回start+1作为结束位置（如果start就是最后一个字符，返回len(expr)）
	if start == len(expr)-1 {
		return len(expr)
	}
	return start + 1
}

// AddUsernameControlToPacketRatePanel 为Packet Rate面板添加username变量控制逻辑
// 当选择特定用户时，只显示该用户的数据；当选择"All"时，显示总体统计
// 实现方式：修改查询表达式，使用or操作符实现条件显示
// 当username="All"（即.*）时，总体查询显示，用户查询返回空
// 当选择特定用户时，总体查询返回空，用户查询正常显示
func AddUsernameControlToPacketRatePanel(dashboard map[string]interface{}) error {
	panels, ok := dashboard["panels"].([]interface{})
	if !ok {
		return nil
	}

	for _, p := range panels {
		panel, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// 只处理Packet Rate面板
		title, ok := panel["title"].(string)
		if !ok || title != "Packet Rate" {
			continue
		}

		targets, ok := panel["targets"].([]interface{})
		if !ok {
			continue
		}

		// 修改查询表达式，使其根据username变量动态显示
		for _, t := range targets {
			target, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			refId, ok := target["refId"].(string)
			if !ok {
				continue
			}

			// refId A和B是总体统计（RX/TX Packets Total）
			// 当username="All"（即.*）时显示，选择特定用户时隐藏
			// 实现方式：使用count检查用户查询匹配的用户数
			// 当username="All"（.*）时，count等于总用户数，显示总体
			// 当选择特定用户时，count=1（或小于总用户数），隐藏总体
			// 技巧：当count等于总用户数时（即选择了All），显示总体；否则返回0

			switch refId {
			case "A":
				// RX Packets Total：当用户查询匹配的用户数等于总用户数时显示，否则返回0
				// 使用count检查：count(用户查询) == count(所有用户) ? 显示总体 : 返回0
				// 总用户数通过count(count(net_user_rx_packets) by (username))获取
				newExpr := "(count(sum(rate(net_user_rx_packets{username=~\"$username\"}[1m])) by (username)) == count(count(net_user_rx_packets) by (username))) * rate(net_interface_rx_packets{iface=\"wg0\"}[1m])"
				target["expr"] = newExpr
			case "B":
				// TX Packets Total：当用户查询匹配的用户数等于总用户数时显示，否则返回0
				newExpr := "(count(sum(rate(net_user_tx_packets{username=~\"$username\"}[1m])) by (username)) == count(count(net_user_tx_packets) by (username))) * rate(net_interface_tx_packets{iface=\"wg0\"}[1m])"
				target["expr"] = newExpr
			}
			// refId C和D（用户查询）保持不变，正常显示
		}
	}

	return nil
}
