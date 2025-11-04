package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tunnel-monitor/internal/config"
)

// ImportDashboard 将 dashboard 导入到 Grafana
func ImportDashboard(dashboard map[string]interface{}) error {
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

// GetDashboards 从 Grafana 获取所有 dashboard 列表
func GetDashboards() ([]map[string]interface{}, error) {
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
