package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func InstallAll() error {
	fmt.Println("🔧 开始安装监控组件...")

	if err := InstallPrometheus(); err != nil {
		return fmt.Errorf("安装 Prometheus 失败: %w", err)
	}

	if err := InstallGrafana(); err != nil {
		return fmt.Errorf("安装 Grafana 失败: %w", err)
	}

	fmt.Println("✅ 所有组件安装完成")
	return nil
}

func InstallPrometheus() error {
	fmt.Println("📦 检查 Prometheus...")

	// 检查是否已安装
	if isCommandAvailable("prometheus") {
		fmt.Println("✅ Prometheus 已安装")
		return nil
	}

	fmt.Println("📥 安装 Prometheus...")

	switch runtime.GOOS {
	case "linux":
		return installPrometheusLinux()
	case "darwin":
		return installPrometheusMacOS()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func InstallGrafana() error {
	fmt.Println("📦 检查 Grafana...")

	// 检查是否已安装
	if isCommandAvailable("grafana-server") || fileExists("/usr/sbin/grafana-server") {
		fmt.Println("✅ Grafana 已安装")
		return nil
	}

	fmt.Println("📥 安装 Grafana...")

	switch runtime.GOOS {
	case "linux":
		return installGrafanaLinux()
	case "darwin":
		return installGrafanaMacOS()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func installPrometheusLinux() error {
	// 检测发行版
	distro := detectLinuxDistro()

	if distro == "debian" || distro == "ubuntu" {
		return installPrometheusDebian()
	} else if distro == "redhat" || distro == "centos" {
		return installPrometheusRedHat()
	}

	return fmt.Errorf("不支持的 Linux 发行版: %s", distro)
}

func installPrometheusDebian() error {
	version := "2.48.0"
	arch := "linux-amd64"
	url := fmt.Sprintf("https://github.com/prometheus/prometheus/releases/download/v%s/prometheus-%s.%s.tar.gz", version, version, arch)

	// 下载并安装
	cmd := exec.Command("bash", "-c", fmt.Sprintf(`
		cd /tmp &&
		wget %s &&
		tar -xzf prometheus-%s.%s.tar.gz &&
		sudo mv prometheus-%s.%s/prometheus /usr/local/bin/ &&
		sudo chmod +x /usr/local/bin/prometheus &&
		rm -rf prometheus-*
	`, url, version, arch, version, arch))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installPrometheusRedHat() error {
	// 与 Debian 相同的方式
	return installPrometheusDebian()
}

func installGrafanaLinux() error {
	distro := detectLinuxDistro()

	if distro == "debian" || distro == "ubuntu" {
		return installGrafanaDebian()
	} else if distro == "redhat" || distro == "centos" {
		return installGrafanaRedHat()
	}

	return fmt.Errorf("不支持的 Linux 发行版: %s", distro)
}

func installGrafanaDebian() error {
	cmd := exec.Command("bash", "-c", `
		if [ ! -f /etc/apt/sources.list.d/grafana.list ]; then
			curl -fsSL https://apt.grafana.com/gpg.key | sudo apt-key add -
			echo "deb https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
			sudo apt update
		fi
		sudo apt install -y grafana
	`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installGrafanaRedHat() error {
	cmd := exec.Command("bash", "-c", `
		if [ ! -f /etc/yum.repos.d/grafana.repo ]; then
			sudo tee /etc/yum.repos.d/grafana.repo > /dev/null <<EOF
[grafana]
name=Grafana
baseurl=https://packages.grafana.com/oss/rpm
enabled=1
gpgcheck=1
gpgkey=https://packages.grafana.com/gpg.key
EOF
			sudo yum makecache
		fi
		sudo yum install -y grafana
	`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installPrometheusMacOS() error {
	cmd := exec.Command("brew", "install", "prometheus")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installGrafanaMacOS() error {
	cmd := exec.Command("brew", "install", "grafana")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func detectLinuxDistro() string {
	// 检查 /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}

	content := string(data)
	if strings.Contains(content, "Ubuntu") || strings.Contains(content, "Debian") {
		return "debian"
	}
	if strings.Contains(content, "CentOS") || strings.Contains(content, "Red Hat") || strings.Contains(content, "Rocky") {
		return "redhat"
	}

	return "unknown"
}
