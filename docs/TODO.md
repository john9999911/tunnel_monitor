# TODO for monitor


## 全局变量
监控面板的全局变量是带宽线路，我想要使用带宽线路查询关联的POP机和客户机，通过数据库来查询。
1. 首先要了解业务，通过服务端的文档和sql脚本来了解；
2. 分析关联：一个带宽线路，它关联的POP机和用户机如何拿到；
3. 全局变量选择一个带宽线路后，如何让各个查询能够使用到该带宽线路关联的数据（POP机等）。
4. 目前监控项目已经部分实现了，需要查看代码，了解进展。



## ✅ 已完成

### 1. 创建全局变量：带宽线路及关联机器
- [x] 在 `business-base.json` 和 `business-template.json` 中添加全局变量定义
- [x] **主变量**：`$bandwidth_line` - 从MySQL `bandwidth_lines` 表查询活跃线路，支持 "All" 选项
- [x] **派生变量**：`$pop_machines` - 基于选择的带宽线路，查询关联的POP机器intra_ip列表
- [x] **派生变量**：`$user_machines` - 基于选择的带宽线路，查询关联的用户机器intra_ip列表
- [x] **修复变量配置**：添加 `allValue: ".*"` 和正确的初始值，确保变量能正确解析和匹配（参见 [VARIABLE_FIX.md](./VARIABLE_FIX.md)）

### 2. 数据库关联查询实现
- [x] **POP机器查询**：`bandwidth_lines` JOIN `machines` WHERE type='pop' AND bandwidth_line_code匹配
- [x] **用户机器查询**：`configs` JOIN `machines` WHERE user_machine_code匹配 AND bandwidth_line_code匹配
- [x] 支持"All"选项：当选择"All"时，查询所有活跃线路关联的机器

### 3. 面板筛选实现
- [x] **客户端面板**（12个）：Prometheus查询使用 `exported_instance=~"$pop_machines"` 或 `user_machine_ip=~"$user_machines"`
- [x] **服务端面板**（8个）：Prometheus查询使用 `pop_intra_ip=~"$pop_machines"`，MySQL查询使用 `bandwidth_line_code REGEXP '$bandwidth_line'`
- [x] 动态筛选：选择不同带宽线路时，变量自动刷新，面板显示相应机器的数据

### 4. 筛选效果验证
- [x] **All选项**：展示所有活跃带宽线路关联的机器指标
- [x] **单线路选择**：仅显示该线路关联的POP机器和用户机器的指标
- [x] **性能优化**：通过数据库预查询机器列表，避免在Prometheus中进行复杂的标签匹配

## 📝 注意事项

### 数据标签要求
客户端和服务端上报的Prometheus指标必须包含相应的机器IP标签，以便与数据库中的intra_ip匹配：

```promql
# 客户端指标示例（POP机器）
pop_alive_status{exported_instance="10.0.0.1", bandwidth_line_code="line001"}

# 客户端指标示例（用户机器）
pop_traffic_tx_rate{user_machine_ip="192.168.1.10", bandwidth_line_code="line001"}

# 服务端指标示例
server_pop_latency{pop_intra_ip="10.0.0.1", bandwidth_line_code="line001"}
```

### 数据库表结构要求
- `bandwidth_lines` 表：包含 `bandwidth_line_code`, `machine_a_code`, `machine_b_code`, `is_active`, `deleted_at`
- `machines` 表：包含 `machine_code`, `intra_ip`, `type` ('pop'/'user'), `deleted_at`
- `configs` 表：包含 `bandwidth_line_code`, `user_machine_code`, `status`, `deleted_at`

### 变量刷新机制
- 所有变量设置为 `refresh=1`，从数据库实时查询
- 选择 `$bandwidth_line` 后，`$pop_machines` 和 `$user_machines` 自动刷新
- 支持正则匹配：`REGEXP '$bandwidth_line'` 用于MySQL查询的"All"选项支持
- **关键配置**：multi-value变量必须设置 `allValue: ".*"` 以支持"All"选项的正则匹配

## 🔧 最近修复

### MySQL查询SQL语法错误（2025-12-23）

**问题**：面板"各线路带宽总量"和"各线路带宽被购买使用情况"报错
```
Error 1064: You have an error in your SQL syntax near 'All GROUP BY bandwidth_line_code'
```

**根本原因**：
- SQL使用了 `REGEXP ${bandwidth_line:singlequote}`
- 当 `$bandwidth_line = 'All'` 时，变成 `REGEXP 'All'`
- `'All'` 不是有效的正则表达式，导致SQL语法错误

**修复方案**：
将 `REGEXP` 改为条件判断：
```sql
-- 修复前（错误）
WHERE ... AND bandwidth_line_code REGEXP ${bandwidth_line:singlequote}

-- 修复后（正确）
WHERE ... AND (${bandwidth_line:singlequote} = 'All' OR bandwidth_line_code = ${bandwidth_line:singlequote})
```

**工作原理**：
- 选择"All"时：`('All' = 'All' OR ...)` → 第一个条件为true → 返回所有数据
- 选择特定线路时：`('line001' = 'All' OR bandwidth_line_code = 'line001')` → 第二个条件匹配 → 返回特定数据

**修复的面板**：
1. 各线路带宽总量.json - 2个查询
2. 各线路带宽被购买使用情况.json - 1个查询

---

### 变量查询返回空值问题（2025-12-23）

**问题**：`$pop_machines` 和 `$user_machines` 在Grafana中显示为空，无法筛选数据

**根本原因分析**：
1. **主变量缺少 `allValue`**：`bandwidth_line` 变量没有设置 `allValue: "All"`
2. **派生变量查询失败**：当 `$bandwidth_line` 为undefined时，SQL条件无法正确匹配
   - SQL条件：`(bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')`
   - 如果 `$bandwidth_line` 为空，条件变成：`(bl.bandwidth_line_code = '' OR '' = 'All')` → 永远为false
3. **数据库查询本身正常**：通过直接测试SQL确认数据存在且查询逻辑正确

**验证方法**：
- ✅ 数据库连接测试：`mysql -u root tunnel -e "SELECT ..."`
- ✅ SQL查询测试：所有查询返回正确数据
  - 带宽线路：1条
  - POP机器：2个（10.1.0.2, 10.1.0.3）
  - 用户机器：2个（10.1.101.2, 10.1.101.3）
- ✅ All选项测试：SQL逻辑正确处理"All"值

**修复措施**：
1. 为 `bandwidth_line` 添加 `allValue: "All"`
2. 确保主变量能正确传递值给派生变量
3. 派生变量已有 `allValue: ".*"` 用于Prometheus正则匹配

**测试脚本**：
- `scripts/verify_variables.sh` - 验证JSON配置
- `scripts/test_variable_queries.sh` - 测试SQL查询

**详细说明**：参见 [VARIABLE_FIX.md](./VARIABLE_FIX.md)

---

### 变量无法解析问题（2025-12-22）
**问题**：`$pop_machines` 和 `$user_machines` 在面板查询中显示为空

**根本原因**：
- 缺少 `allValue` 字段：当选择"All"时，变量无法生成有效的正则表达式
- 初始值为空字符串：导致变量初始化失败
- `includeAll` 设置为 `false`：限制了筛选灵活性

**修复措施**：
1. 添加 `allValue: ".*"` 到 `pop_machines` 和 `user_machines` 变量
2. 设置初始值为 `value: "$__all"`, `text: "All"`
3. 修改 `includeAll: true` 以支持"All"选项
4. 确保 `current.selected: true` 以激活初始值

**效果**：
- 选择"All"时：变量值为 `.*`，匹配所有机器
- 选择特定线路时：变量值为实际IP列表，精确筛选
- 变量值始终有效，不会出现空值导致查询失败

**详细说明**：参见 [VARIABLE_FIX.md](./VARIABLE_FIX.md)
