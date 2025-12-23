# Tunnel Monitor

独立的监控面板管理工具，用于管理 Prometheus 和 Grafana。

## 功能特性

- ✅ 自动安装 Prometheus 和 Grafana
- ✅ 启动和停止监控服务
- ✅ 创建和管理业务监控面板
- ✅ 支持按带宽线路筛选监控数据
- ✅ 统一展示客户端和服务端指标
- ✅ 客户端监控数据由服务端转发，无需主动发现客户端
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
- 业务监控面板模板：`./dashboards/business-template.json`
- 面板组件：`./dashboards/panels/client/` 和 `./dashboards/panels/server/`

如果需要自定义配置，可以编辑 `config.yaml` 或使用 `--config` 参数指定配置文件。

**重要配置项**：

### MySQL数据源配置

面板中的部分指标需要查询MySQL数据库（如带宽线路筛选、订单统计等）。需要配置MySQL数据源：

```yaml
mysql:
  host: "localhost"
  port: 3306
  database: "iptunnel"
  username: "root"
  password: "your_password"
  uid: "mysql-datasource"  # 必须与Grafana中MySQL数据源的UID一致
```

**配置步骤**：
1. 在Grafana中创建MySQL数据源（Configuration → Data sources → Add MySQL）
2. 配置连接信息并保存
3. 记录数据源的UID（在URL中可以看到）
4. 将该UID填入 `config.yaml` 的 `mysql.uid` 字段
5. 重新创建面板：`./tunnel-monitor dashboard create`

详细说明请参考：[MySQL数据源配置文档](docs/MYSQL_DATASOURCE.md)

**注意**：客户端监控数据现在由服务端转发到Prometheus，无需配置客户端列表。

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
# 创建业务监控面板（推荐 - 包含所有指标）
./tunnel-monitor dashboard create

# 或者创建所有面板（同上）
./tunnel-monitor dashboard create-all

# 列出所有面板
./tunnel-monitor dashboard list
```

**面板特性**：
- 统一展示客户端和服务端指标
- 支持按带宽线路筛选（选择"All"显示所有线路）
- 客户端数据通过`exported_instance`标签区分不同的POP机器
- 包含流量监控、延迟监控、状态监控、带宽分配等所有业务指标

## 完整工作流程

```bash
# 1. 安装监控组件
sudo ./tunnel-monitor install

# 2. 更新 Prometheus 配置（指向服务端metrics地址）
./tunnel-monitor prometheus update-config

# 3. 启动监控服务
sudo ./tunnel-monitor start

# 4. 创建业务监控面板
./tunnel-monitor dashboard create

# 5. 访问 Grafana (默认: http://localhost:3000)
#    用户名: admin
#    密码: admin
```

**监控架构说明**：
- 客户端(POP)指标由服务端收集并转发到Prometheus
- 服务端自身指标直接暴露给Prometheus
- 所有指标统一在业务监控面板中展示
- 使用`bandwidth_line_code`标签实现线路级别筛选
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

