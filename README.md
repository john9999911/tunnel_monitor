# Tunnel Monitor

独立的监控面板管理工具，用于管理 Prometheus 和 Grafana。

## 功能特性

- ✅ 自动安装 Prometheus 和 Grafana
- ✅ 启动和停止监控服务
- ✅ 创建和管理监控面板
- ✅ 配置服务端和客户端监控 URL
- ✅ 支持统一监控面板（使用变量选择客户端实例）
- ✅ 独立于 tunnel_server，可单独使用

## 安装

```bash
cd tunnel_monitor
go mod tidy
go build -o tunnel-monitor
```

## 配置

配置文件 `config.yaml` 已经配置好，所有配置文件都在项目内部：

- Prometheus 配置：`./config/monitoring/prometheus.yml`
- 服务端面板模板：`./dashboards/server-template.json`
- 客户端面板模板：`./dashboards/client-template.json`

如果需要自定义配置，可以编辑 `config.yaml` 或使用 `--config` 参数指定配置文件。

## 使用方法

### 安装监控组件

```bash
# 安装 Prometheus 和 Grafana
./tunnel-monitor install
```

### 启动监控服务

```bash
# 启动 Prometheus 和 Grafana
./tunnel-monitor start
```

### 停止监控服务

```bash
./tunnel-monitor stop
```

### 查看服务状态

```bash
./tunnel-monitor status
```

### 更新 Prometheus 配置

```bash
# 根据 config.yaml 中的客户端配置更新 Prometheus 配置
./tunnel-monitor prometheus update-config
```

### 创建监控面板

```bash
# 创建统一客户端监控面板（支持多客户端，使用变量选择）
./tunnel-monitor dashboard create-unified

# 创建统一服务端监控面板（支持多服务端部署，使用变量选择）
./tunnel-monitor dashboard create-server

# 创建数据库监控面板
./tunnel-monitor dashboard create-database

# 列出所有面板
./tunnel-monitor dashboard list
```

## 完整工作流程

```bash
# 1. 安装监控组件
sudo ./tunnel-monitor install

# 2. 配置客户端（编辑 config.yaml）
vim config.yaml

# 3. 更新 Prometheus 配置
./tunnel-monitor prometheus update-config

# 4. 启动监控服务
sudo ./tunnel-monitor start

# 5. 创建监控面板
./tunnel-monitor dashboard create-server      # 统一服务端监控面板
./tunnel-monitor dashboard create-unified     # 统一客户端监控面板
./tunnel-monitor dashboard create-database    # 数据库监控面板
```

## 项目结构

```
tunnel_monitor/
├── cmd/                    # CLI 命令
│   ├── root.go
│   ├── install.go
│   ├── start.go
│   ├── stop.go
│   ├── status.go
│   ├── dashboard.go
│   └── prometheus.go
├── internal/
│   ├── config/            # 配置管理
│   ├── installer/         # 安装器
│   ├── service/           # 服务管理
│   ├── dashboard/         # 面板管理
│   └── prometheus/        # Prometheus 配置
├── config/
│   └── monitoring/
│       └── prometheus.yml # Prometheus 配置文件
├── dashboards/
│   ├── server-template.json    # 服务端监控面板模板
│   ├── client-template.json    # 客户端监控面板模板
│   └── database-template.json  # 数据库监控面板模板
├── config.yaml            # 配置文件
├── config.yaml.example    # 配置文件示例
├── go.mod
├── go.sum
└── main.go
```

## 配置说明

### Prometheus 配置

- `url`: Prometheus Web UI 地址
- `port`: Prometheus 端口
- `data_dir`: 数据存储目录
- `config_file`: 配置文件路径

### Grafana 配置

- `url`: Grafana Web UI 地址
- `port`: Grafana 端口
- `username`: Grafana 用户名
- `password`: Grafana 密码
- `api_key`: API Key（可选，用于认证）

### 客户端配置

在 `clients` 列表中配置所有要监控的客户端：

```yaml
clients:
  - name: "客户端名称"
    instance: "IP:PORT"        # 客户端实例地址
    metrics_url: "http://..."  # Metrics 端点 URL
```

## 面板模板

面板模板文件位于 `dashboards/` 目录下：

- `server-template.json`: 服务端监控面板模板
- `client-template.json`: 客户端监控面板模板
- `database-template.json`: 数据库监控面板模板

这些模板文件已经包含在项目中，可以直接使用。

### 统一监控面板特性

**统一客户端监控面板**：
- 从 Prometheus 自动发现所有客户端实例
- 添加 instance 变量供选择
- 为所有查询添加 `instance="$instance"` 过滤
- 支持多客户端部署场景

**统一服务端监控面板**：
- 从 Prometheus 自动发现所有服务端实例
- 添加 instance 变量供选择
- 为所有查询添加 `instance="$instance"` 过滤
- 支持多服务端部署场景（参考 multi-server-deployment.md）

**数据库监控面板**：
- 监控数据库连接数
- 监控数据库查询耗时和QPS
- 监控数据库错误统计
- 简单实用的数据库性能监控

## 许可证

MIT

