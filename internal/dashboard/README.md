# Dashboard 模块说明

## 代码结构

本模块已经按照职责拆分为多个文件：

### 核心文件

- **dashboard.go** - 主入口文件，提供创建和管理dashboard的公开API
- **prometheus.go** - Prometheus查询相关逻辑（获取实例列表等）
- **grafana.go** - Grafana API交互（导入dashboard、获取列表等）
- **template.go** - 模板加载逻辑（支持完整模板和拆分模板）
- **template_manager.go** - 模板管理器，用于拆分和管理大型模板文件
- **builder.go** - Dashboard构建逻辑（添加变量、过滤查询、修复数据源等）
- **utils.go** - 工具函数

## 模板文件拆分

对于大型模板文件（如超过1000行的JSON文件），可以使用模板管理器进行拆分：

### 拆分模板文件

模板管理器可以将完整的dashboard JSON文件拆分为：
- **基础模板** - 包含dashboard的元数据（annotations、templating、time等）
- **Panels目录** - 每个panel保存为独立的JSON文件

### 使用方式

1. **自动拆分**（代码支持，需要手动调用）：
```go
tm := dashboard.NewTemplateManager("./dashboards")
err := tm.SplitTemplate(
    "./dashboards/client-template.json",  // 完整模板路径
    "./dashboards/client-base.json",      // 输出基础模板路径
    "./dashboards/panels/client",         // 输出panels目录
)
```

2. **自动加载**：
代码会自动检测是否存在拆分后的panels目录：
- 如果存在 `./dashboards/panels/client` 目录，则使用拆分模板
- 否则使用完整的模板文件

### 目录结构示例

```
dashboards/
├── client-template.json          # 完整模板（兼容）
├── client-base.json              # 基础模板（拆分后）
├── server-template.json
├── server-base.json
├── database-template.json
├── database-base.json
└── panels/                       # Panels目录
    ├── client/                   # 客户端panels
    │   ├── WireGuard_Status.json
    │   ├── Connection_Stats.json
    │   └── ...
    ├── server/                   # 服务端panels
    │   ├── HTTP_请求率.json
    │   └── ...
    └── database/                 # 数据库panels
        ├── 数据库连接数.json
        └── ...
```

## 优势

1. **代码模块化** - 每个文件职责单一，易于维护
2. **模板可管理** - 大型模板文件可以拆分为小片段，便于版本控制和修改
3. **向后兼容** - 仍然支持完整模板文件，拆分是可选的
4. **灵活扩展** - 可以轻松添加新的panel而不需要修改大型JSON文件

## API 说明

### 主要函数

- `CreateUnifiedDashboard()` - 创建统一客户端监控面板
- `CreateServerDashboard()` - 创建统一服务端监控面板
- `CreateDatabaseDashboard()` - 创建数据库监控面板
- `ListDashboards()` - 列出所有监控面板

### 内部模块

- `GetClientInstances()` - 获取客户端实例列表
- `GetServerInstances()` - 获取服务端实例列表
- `LoadClientTemplate()` - 加载客户端模板
- `LoadServerTemplate()` - 加载服务端模板
- `LoadDatabaseTemplate()` - 加载数据库模板
- `ImportDashboard()` - 导入dashboard到Grafana
- `GetDashboards()` - 获取dashboard列表
