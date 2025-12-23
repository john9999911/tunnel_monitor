# Grafana变量配置修复说明

## 问题描述

用户反馈：面板中的 `$pop_machines` 和 `$user_machines` 变量在查询中显示为空，无法正确筛选数据。

## 根本原因

Grafana的变量配置不完整，缺少关键字段：

1. **主变量缺少 `allValue`**：`bandwidth_line` 变量没有设置 `allValue`，导致选择"All"时传递给派生变量的值无效
2. **派生变量无法获取数据**：当 `$bandwidth_line` 为空或无效时，SQL条件 `(bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')` 无法正确匹配
3. **multi-value变量配置不完整**：派生变量虽然有 `allValue: ".*"`，但依赖主变量提供有效值

## 修复方案

### 1. 为主变量添加 `allValue`

```json
{
  "name": "bandwidth_line",
  "includeAll": true,
  "allValue": "All"  // ← 关键：当选择"All"时，传递字符串"All"给派生变量
}
```

**作用**：
- 当用户选择"All"时，`$bandwidth_line` 被替换为 `"All"`
- 派生变量的SQL条件：`(bl.bandwidth_line_code = 'All' OR 'All' = 'All')` → 第二个条件永远为true
- 这样派生变量就能返回所有数据

### 2. 为派生变量添加 `allValue`

```json
{
  "name": "pop_machines",
  "multi": true,
  "includeAll": true,
  "allValue": ".*"  // ← 关键：匹配所有值的正则表达式
}
```

**作用**：
- 当派生变量查询返回所有机器IP后，如果选择"All"，Grafana将变量替换为 `.*`
- 在Prometheus查询中：`exported_instance=~".*"` 匹配所有实例

### 3. 修正初始值

```json
{
  "current": {
    "selected": true,
    "text": "All",
    "value": "$__all"  // ← 使用Grafana的$__all特殊值
  }
}
```

**作用**：
- 确保变量初始化时有有效值
- `$__all` 是Grafana的特殊标记，表示"所有选项"
- 触发 `allValue` 的使用

### 3. 修正初始值

```json
{
  "current": {
    "selected": true,      // ← 修改：激活初始值
    "text": "All",         // ← 修改：显示文本
    "value": "$__all"      // ← 修改：使用Grafana特殊值
  }
}
```

**作用**：
- 确保变量初始化时有有效值
- `$__all` 是Grafana的特殊标记，表示"所有选项"
- 触发 `allValue` 的使用

### 4. 变量级联的工作流程

```
用户选择带宽线路
    ↓
bandwidth_line = "BANDWIDTH_LINE_xxx" 或 "All"
    ↓
派生变量查询触发
    ↓
pop_machines查询:
  WHERE ... AND (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')
  
  如果$bandwidth_line = "All":
    → WHERE ... AND (bl.bandwidth_line_code = 'All' OR 'All' = 'All')
    → 第二个条件为true，返回所有POP机器
  
  如果$bandwidth_line = "BANDWIDTH_LINE_xxx":
    → WHERE ... AND (bl.bandwidth_line_code = 'BANDWIDTH_LINE_xxx' OR 'BANDWIDTH_LINE_xxx' = 'All')
    → 第一个条件匹配，返回特定线路的POP机器
    ↓
pop_machines = ["10.1.0.2", "10.1.0.3"]
    ↓
面板Prometheus查询:
  pop_alive_status{exported_instance=~"10.1.0.2|10.1.0.3"}
```

### 5. 变量刷新机制

```json
{
  "refresh": 1,  // 1=on dashboard load, 2=on time range change
  "query": "SELECT ... WHERE ... AND (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')"
}
```

**工作流程**：
1. 用户选择 `$bandwidth_line`（例如："line001"）
2. `$pop_machines` 查询自动触发，传入 `'line001'`
3. 查询返回该线路关联的POP机器IP列表
4. `$user_machines` 查询也触发，返回用户机器IP列表
5. 面板的Prometheus查询使用这些IP列表进行筛选

## 修复前后对比

### 修复前 ❌

```json
{
  "name": "pop_machines",
  "current": {
    "text": "",
    "value": ""  // 空值
  },
  "includeAll": false,  // 不支持All选项
  // 缺少 allValue 字段
}
```

**问题**：
- 初始值为空，导致面板查询失败
- 不支持"All"选项，灵活性差
- 变量值可能为undefined

### 修复后 ✅

```json
{
  "name": "pop_machines",
  "current": {
    "text": "All",
    "value": "$__all"  // 有效值
  },
  "includeAll": true,  // 支持All选项
  "allValue": ".*"     // All选项的实际值
}
```

**效果**：
- 初始化时自动显示所有数据
- 支持筛选特定线路或查看全部
- 变量值总是有效的正则表达式

### MySQL查询中的变量

MySQL面板使用条件判断而不是REGEXP：

```sql
-- 正确：使用条件判断
SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL 
  AND ('${bandwidth_line}' = 'All' OR bandwidth_line_code = '${bandwidth_line}')

-- 错误：使用REGEXP（当$bandwidth_line='All'时会失败）
SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL 
  AND bandwidth_line_code REGEXP '${bandwidth_line}'
```

**原因**：
- `REGEXP 'All'` 不是有效的正则表达式匹配
- 条件判断 `'All' = 'All'` 永远为true，返回所有数据
- 条件判断 `'line001' = 'All'` 为false，使用 `bandwidth_line_code = 'line001'` 精确匹配

## 使用示例

### Prometheus查询中的变量

```promql
# POP机器存活状态
pop_alive_status{exported_instance=~"$pop_machines"}

# 当 $bandwidth_line="All" 时：
# → pop_alive_status{exported_instance=~".*"}  # 匹配所有

# 当 $bandwidth_line="line001" 时，假设查询返回 "10.1.0.1,10.1.0.2"
# → pop_alive_status{exported_instance=~"10.1.0.1|10.1.0.2"}

# 用户机器流量
pop_traffic_tx_rate{user_machine_ip=~"$user_machines"}
```

### MySQL查询中的变量

```sql
-- 带宽总量查询
SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 
  AND deleted_at IS NULL 
  AND bandwidth_line_code REGEXP '$bandwidth_line'

-- 当 $bandwidth_line="All" 时：
-- → bandwidth_line_code REGEXP '.*'  # 匹配所有

-- 当 $bandwidth_line="line001" 时：
-- → bandwidth_line_code REGEXP 'line001'  # 精确匹配
```

## 变量格式化

Grafana在将multi-value变量插入查询时，会根据上下文自动格式化：

| 场景 | 格式 | 示例 |
|------|------|------|
| Prometheus正则 | 竖线分隔 | `10.1.0.1\|10.1.0.2` |
| MySQL IN | 逗号分隔加引号 | `'10.1.0.1','10.1.0.2'` |
| 单值选择 | 直接值 | `10.1.0.1` |
| All选项 | allValue | `.*` |

## 验证步骤

1. **检查变量值**：
   - 在Grafana面板设置中，进入"Variables"
   - 查看 `pop_machines` 和 `user_machines` 的"Preview of values"
   - 应该显示实际的IP地址列表

2. **测试筛选**：
   - 选择"All"：应显示所有线路的数据
   - 选择特定线路（如"line001"）：应只显示该线路的数据
   - 切换线路：面板应实时更新

3. **查看查询日志**：
   - Grafana Inspector → Query
   - 查看实际发送的查询语句
   - 确认变量已被正确替换

## 注意事项

### 数据库必须有数据

变量查询依赖数据库：
- `bandwidth_lines` 表必须有活跃线路（`is_active=1`, `deleted_at IS NULL`）
- `machines` 表必须有关联的机器
- `configs` 表必须有活跃订单（`status='active'`）

如果这些表为空，变量将没有可选值。

### MySQL数据源配置

确保 `config.yaml` 中的MySQL UID正确：

```yaml
mysql:
  uid: "df7uceqwaqzggc"  # 必须与Grafana中实际的数据源UID一致
```

### 变量顺序很重要

定义顺序：
1. `bandwidth_line`（主变量）
2. `pop_machines`（依赖主变量）
3. `user_machines`（依赖主变量）

Grafana按顺序刷新变量，确保依赖关系正确。

## 常见问题排查

### Q1: 变量显示为空

**检查**：
- 数据库中是否有数据？
- MySQL数据源配置是否正确？
- SQL查询语法是否正确？

**解决**：
```bash
# 手动测试查询
mysql -u root -p -e "
SELECT DISTINCT m.intra_ip 
FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) 
WHERE bl.is_active=1 AND bl.deleted_at IS NULL AND m.deleted_at IS NULL AND m.type='pop'
"
```

### Q2: 选择"All"后面板无数据

**原因**：`allValue` 设置不正确

**检查**：
- `allValue: ".*"` 是否设置？
- Prometheus查询是否使用 `=~` 正则匹配？
- MySQL查询是否使用 `REGEXP`？

### Q3: 切换线路后面板不更新

**原因**：变量刷新配置问题

**检查**：
- `refresh: 1` 是否设置？
- 派生变量的查询是否包含 `$bandwidth_line`？

## 参考文档

- [Grafana Variables Documentation](https://grafana.com/docs/grafana/latest/variables/)
- [Prometheus Query Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [tunnel_monitor/docs/FILTERING_STRATEGY.md](./FILTERING_STRATEGY.md)
