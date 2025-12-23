# Dashboard 模板文件说明

## 文件结构

模板文件已经拆分为更易管理的结构：

```
dashboards/
├── business-base.json            # 业务监控基础模板（不含panels）
├── business-template.json        # 业务监控完整模板
└── panels/                       # Panels目录
    ├── client/                   # 客户端panels（12个文件）
    │   ├── POP客户端软件版本号.json
    │   ├── WireGuard_peer连接数.json
    │   ├── 各用户机器发送流量速率.json
    │   ├── 各用户机器接收流量速率.json
    │   ├── 各带宽线路对端网络延迟.json
    │   ├── 与各peer节点连接状态.json
    │   ├── POP客户端存活状态.json
    │   ├── 域名解析服务状态.json
    │   ├── 用户上传限速触发.json
    │   ├── 用户下载限速触发.json
    │   ├── 每个用户机器的发送字节数.json
    │   └── 每个用户机器的接收字节数.json
    └── server/                   # 服务端panels（8个文件）
        ├── 服务端软件版本号.json
        ├── 服务端健康状态.json
        ├── 各用户订单个数.json
        ├── 业务统计.json
        ├── 服务端到POP延迟.json
        ├── POP端与服务端通信状态.json
        ├── 各线路带宽总量.json
        └── 各线路带宽被购买使用情况.json
```

## 全局变量

### 带宽线路筛选
- **主变量**：`$bandwidth_line` - 从数据库查询活跃的带宽线路
  - 数据源：MySQL数据库 `bandwidth_lines` 表
  - 支持"All"选项显示所有线路
  
- **派生变量**（自动刷新）：
  - `$pop_machines` - 根据选择的带宽线路，查询关联的POP机器IP列表
  - `$user_machines` - 根据选择的带宽线路，查询关联的用户机器IP列表

### 重要配置

所有multi-value变量必须包含以下字段：
```json
{
  "includeAll": true,     // 支持"All"选项
  "allValue": ".*",       // "All"选项的实际值（正则表达式）
  "multi": true,          // 允许多选
  "refresh": 1,           // 自动刷新
  "current": {
    "value": "$__all",    // 初始值
    "text": "All",
    "selected": true
  }
}
```

**主变量（bandwidth_line）的allValue**：
- 设置为 `"All"` 而不是 `".*"`
- 原因：派生变量的SQL使用条件判断 `('$bandwidth_line' = 'All' OR ...)`
- 如果设置为 `".*"`，条件会变成 `('.*' = 'All' OR ...)`，第一个条件永远为false

**派生变量的allValue**：
- 设置为 `".*"` 用于Prometheus正则匹配
- 当选择"All"时，Prometheus查询变成 `exported_instance=~".*"`，匹配所有实例

### MySQL查询注意事项

**正确的SQL模式**：
```sql
-- 使用条件判断，而不是REGEXP
WHERE ... AND ('${bandwidth_line}' = 'All' OR bandwidth_line_code = '${bandwidth_line}')
```

**错误的SQL模式**（会导致语法错误）：
```sql
-- ❌ 不要使用REGEXP，当$bandwidth_line='All'时会失败
WHERE ... AND bandwidth_line_code REGEXP '${bandwidth_line}'
```

**为什么需要条件判断？**
- 当 `$bandwidth_line = "All"` 时：`REGEXP 'All'` 导致SQL语法错误
- 使用条件判断：`('All' = 'All' OR ...)` → 第一个条件为true → 返回所有数据

### 为什么需要 `allValue`？
- 当选择"All"时，Grafana将变量替换为 `.*`
- 在Prometheus查询中：`exported_instance=~".*"` 匹配所有实例
- 在MySQL查询中：`REGEXP '.*'` 匹配所有行
- 没有这个字段，变量将为空，导致查询失败

详细说明：参见 [docs/VARIABLE_FIX.md](../docs/VARIABLE_FIX.md)

## 面板特性

### 全局变量
- **带宽线路筛选**：通过全局变量 `$bandwidth_line` 实现按带宽线路筛选
  - 支持 "All" 选项，显示所有线路数据
  - 选择特定线路后，仅显示该线路相关的：
    - POP机器指标
    - 线路上的用户流量
    - 线路带宽使用情况

### 面板结构
- **客户端指标**：12个panel，监控POP客户端状态、流量、延迟等
- **服务端指标**：8个panel，监控服务端健康、业务数据、带宽分配等
- **总计**：20个panel文件，每个文件约50-200行

## 使用方式

### 修改Panel

可以直接编辑单个panel文件，无需修改巨大的JSON文件：

```bash
# 编辑客户端某个panel
vim dashboards/panels/client/各用户机器发送流量速率.json

# 编辑服务端某个panel
vim dashboards/panels/server/服务端到POP延迟.json
```

### 全局变量使用

所有面板中的Prometheus查询都包含带宽线路筛选：

```promql
# 示例：客户端流量查询
pop_traffic_tx_rate{bandwidth_line_code=~"$bandwidth_line"} / 1000
```

MySQL查询也支持带宽线路筛选：

```sql
-- 示例：带宽总量查询
SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL 
AND bandwidth_line_code REGEXP '$bandwidth_line'
```

## 优势

1. **易于维护**：每个panel独立文件，修改不会影响其他panel
2. **版本控制友好**：Git diff可以精确显示哪个panel被修改
3. **代码组织清晰**：基础配置和面板分离，结构更清晰
4. **向后兼容**：仍然支持完整模板文件，拆分是可选的

## 数据源配置

面板模板中使用占位符，部署时会被替换为实际的数据源UID：

- `{{PROMETHEUS_UID}}` - Prometheus数据源
- `{{MYSQL_UID}}` - MySQL数据源

## 注意事项

1. **带宽线路标签**：所有客户端和服务端指标必须包含 `bandwidth_line_code` 标签
2. **数据库表结构**：`bandwidth_lines` 表需要包含 `bandwidth_line_code` 字段用于筛选
3. **变量刷新**：带宽线路变量设置为自动刷新（refresh=1），从数据库实时获取活跃线路
3. **顺序**：Panels按文件名排序加载，如需特定顺序，请使用数字前缀（如 `01_xxx.json`）
4. **备份**：原始的完整模板文件已保留，可作为备份参考
