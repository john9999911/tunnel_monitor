package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"tunnel-monitor/internal/config"
)

func StartAll() error {
	fmt.Println("🚀 启动监控服务...")

	if err := StartPrometheus(); err != nil {
		return fmt.Errorf("启动 Prometheus 失败: %w", err)
	}

	if err := StartGrafana(); err != nil {
		return fmt.Errorf("启动 Grafana 失败: %w", err)
	}

	fmt.Println("✅ 所有服务已启动")
	return nil
}

func StopAll() error {
	fmt.Println("🛑 停止监控服务...")

	if err := StopPrometheus(); err != nil {
		fmt.Printf("⚠️ 停止 Prometheus 失败: %v\n", err)
	}

	if err := StopGrafana(); err != nil {
		fmt.Printf("⚠️ 停止 Grafana 失败: %v\n", err)
	}

	fmt.Println("✅ 服务已停止")
	return nil
}

func StartPrometheus() error {
	// 检查是否已经在运行
	if isProcessRunning("prometheus") {
		fmt.Println("✅ Prometheus 已在运行")
		return nil
	}

	cfg := config.Global

	// 查找 prometheus 可执行文件
	promBin := findPrometheusBinary()
	if promBin == "" {
		return fmt.Errorf("未找到 Prometheus 可执行文件，请先运行 'tunnel-monitor install'")
	}

	// 确保数据目录存在（使用绝对路径或相对于配置文件的路径）
	dataDir := cfg.Prometheus.DataDir
	if !strings.HasPrefix(dataDir, "/") && !strings.HasPrefix(dataDir, "./") {
		// 相对路径，转换为绝对路径
		cwd, _ := os.Getwd()
		dataDir = fmt.Sprintf("%s/%s", cwd, dataDir)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 启动 Prometheus
	configFile := cfg.Prometheus.ConfigFile
	if !strings.HasPrefix(configFile, "/") {
		// 相对路径，转换为绝对路径
		cwd, _ := os.Getwd()
		configFile = fmt.Sprintf("%s/%s", cwd, configFile)
	}

	args := []string{
		"--config.file=" + configFile,
		"--storage.tsdb.path=" + dataDir,
		"--storage.tsdb.retention.time=200h",
		"--web.enable-lifecycle",
	}

	cmd := exec.Command(promBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Prometheus 失败: %w", err)
	}

	fmt.Println("✅ Prometheus 启动成功")
	return nil
}

func StartGrafana() error {
	// 检查是否已经在运行
	if isProcessRunning("grafana-server") {
		fmt.Println("✅ Grafana 已在运行")
		return nil
	}

	// 尝试使用 systemd 启动（Linux）
	if runtime.GOOS == "linux" {
		if err := startGrafanaSystemd(); err == nil {
			fmt.Println("✅ Grafana 启动成功（systemd）")
			return nil
		}
	}

	// 直接启动
	grafBin := findGrafanaBinary()
	if grafBin == "" {
		return fmt.Errorf("未找到 Grafana 可执行文件，请先运行 'tunnel-monitor install'")
	}

	cmd := exec.Command(grafBin, "--config=/etc/grafana/grafana.ini", "--homepath=/usr/share/grafana")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Grafana 失败: %w", err)
	}

	fmt.Println("✅ Grafana 启动成功")
	return nil
}

func StopPrometheus() error {
	if !isProcessRunning("prometheus") {
		return nil
	}

	cmd := exec.Command("pkill", "-f", "prometheus")
	return cmd.Run()
}

func StopGrafana() error {
	// 尝试使用 systemd 停止
	if runtime.GOOS == "linux" {
		if err := stopGrafanaSystemd(); err == nil {
			return nil
		}
	}

	if !isProcessRunning("grafana-server") {
		return nil
	}

	cmd := exec.Command("pkill", "-f", "grafana-server")
	return cmd.Run()
}

func ShowStatus() error {
	fmt.Println("📊 监控服务状态:")
	fmt.Println()

	// Prometheus 状态
	if isProcessRunning("prometheus") {
		fmt.Println("✅ Prometheus: 运行中")
		fmt.Printf("   URL: %s\n", config.Global.Prometheus.URL)
	} else {
		fmt.Println("❌ Prometheus: 未运行")
	}

	// Grafana 状态
	if isProcessRunning("grafana-server") {
		fmt.Println("✅ Grafana: 运行中")
		fmt.Printf("   URL: %s\n", config.Global.Grafana.URL)
		fmt.Printf("   用户名: %s\n", config.Global.Grafana.Username)
	} else {
		fmt.Println("❌ Grafana: 未运行")
	}

	return nil
}

func isProcessRunning(processName string) bool {
	cmd := exec.Command("pgrep", "-f", processName)
	err := cmd.Run()
	return err == nil
}

func findPrometheusBinary() string {
	paths := []string{
		"prometheus",
		"/usr/local/bin/prometheus",
		"/usr/bin/prometheus",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	return ""
}

func findGrafanaBinary() string {
	paths := []string{
		"grafana-server",
		"/usr/sbin/grafana-server",
		"/usr/local/bin/grafana-server",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	return ""
}

func startGrafanaSystemd() error {
	cmd := exec.Command("sudo", "systemctl", "start", "grafana-server")
	return cmd.Run()
}

func stopGrafanaSystemd() error {
	cmd := exec.Command("sudo", "systemctl", "stop", "grafana-server")
	return cmd.Run()
}
