package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tunnel-monitor/internal/config"
)

// GetClientInstances 从 Prometheus targets API 获取所有客户端实例（包括 up 和 down）
func GetClientInstances() ([]string, error) {
	// 从 Prometheus targets API 获取所有配置的 target，包括 down 的
	promURL := config.Global.Prometheus.URL
	url := fmt.Sprintf("%s/api/v1/targets", promURL)

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
	activeTargets := dataObj["activeTargets"].([]interface{})

	instances := make(map[string]bool)
	for _, target := range activeTargets {
		targetMap := target.(map[string]interface{})
		labels := targetMap["labels"].(map[string]interface{})
		job := getString(labels, "job")

		// 获取所有 tunnel-client-pop job 的目标，不管健康状态如何（up/down/unknown）
		if job == "tunnel-client-pop" {
			instance := getString(labels, "instance")
			if instance != "" && strings.Contains(instance, ":") {
				instances[instance] = true
			}
		}
	}

	resultList := make([]string, 0, len(instances))
	for inst := range instances {
		resultList = append(resultList, inst)
	}

	return resultList, nil
}

// GetServerInstances 从 Prometheus 查询服务端实例
func GetServerInstances() ([]string, error) {
	// 方法1: 从 Prometheus targets API 获取
	promURL := config.Global.Prometheus.URL
	url := fmt.Sprintf("%s/api/v1/targets", promURL)

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
	activeTargets := dataObj["activeTargets"].([]interface{})

	instances := make(map[string]bool)
	for _, target := range activeTargets {
		targetMap := target.(map[string]interface{})
		labels := targetMap["labels"].(map[string]interface{})
		job := getString(labels, "job")
		health := getString(targetMap, "health")

		// 只获取服务端 job 且健康的目标
		if job == "tunnel-server" && (health == "up" || health == "unknown") {
			instance := getString(labels, "instance")
			if instance != "" && strings.Contains(instance, ":") {
				instances[instance] = true
			}
		}
	}

	// 方法2: 如果从 targets 没找到，尝试从指标查询
	if len(instances) == 0 {
		query := `tunnel_server_users_total{job="tunnel-server"}`
		encodedQuery := strings.ReplaceAll(query, " ", "%20")
		queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", promURL, encodedQuery)

		resp, err := http.Get(queryURL)
		if err == nil {
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err == nil {
				var queryResult map[string]interface{}
				if json.Unmarshal(data, &queryResult) == nil {
					if getString(queryResult, "status") == "success" {
						dataObj := queryResult["data"].(map[string]interface{})
						results := dataObj["result"].([]interface{})
						for _, r := range results {
							metric := r.(map[string]interface{})["metric"].(map[string]interface{})
							instance := getString(metric, "instance")
							if instance != "" && strings.Contains(instance, ":") {
								instances[instance] = true
							}
						}
					}
				}
			}
		}
	}

	resultList := make([]string, 0, len(instances))
	for inst := range instances {
		resultList = append(resultList, inst)
	}

	return resultList, nil
}
