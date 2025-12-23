# 监控面板筛选策略

## 问题分析

### 当前问题

面板中使用 `bandwidth_line_code=~"$bandwidth_line"` 来筛选Prometheus指标，但存在以下问题：

1. **Prometheus指标不包含 `bandwidth_line_code` 标签**
   - 客户端指标只有：`instance`, `exported_instance`, `user_machine_ip`, `user_code`, `peer_ip` 等
   - 服务端指标只有：`instance`, `pop_ip`, `pop_intra_ip`, `pop_id` 等
   
2. **带宽线路与机器的关系在数据库中**
   - `bandwidth_lines` 表: `bandwidth_line_code` → `machine_a_code`, `machine_b_code`
   - `configs` 表: `bandwidth_line_code` → `user_machine_code`
   - Prometheus无法直接访问这些关系

### 架构分析

```
用户选择带宽线路 ($bandwidth_line)
    ↓
查询数据库获取相关机器
    ├─ POP机器列表 (machine_a_code, machine_b_code)
    └─ 用户机器列表 (user_machine_code from configs)
    ↓
使用机器标识筛选Prometheus指标
    ├─ exported_instance (POP机器的intra_ip)
    └─ user_machine_ip (用户机器的intra_ip)
```

## 解决方案

### 方案1: Grafana变量级联（推荐）✅

**原理**: 使用Grafana的变量级联功能，先查询数据库获取机器列表，再用机器标识筛选Prometheus

**实现步骤**:

#### 1. 定义全局变量

```json
{
  "templating": {
    "list": [
      {
        "name": "bandwidth_line",
        "type": "query",
        "datasource": "MySQL",
        "query": "SELECT DISTINCT bandwidth_line_code FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL",
        "label": "带宽线路",
        "includeAll": true
      },
      {
        "name": "pop_machines",
        "type": "query",
        "datasource": "MySQL",
        "query": "SELECT DISTINCT m.intra_ip FROM bandwidth_lines bl JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) WHERE bl.deleted_at IS NULL AND m.deleted_at IS NULL AND m.type='pop' AND (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')",
        "hide": 2,
        "multi": true
      },
      {
        "name": "user_machines",
        "type": "query",
        "datasource": "MySQL",
        "query": "SELECT DISTINCT m.intra_ip FROM configs c JOIN machines m ON c.user_machine_code = m.machine_code WHERE c.deleted_at IS NULL AND m.deleted_at IS NULL AND c.status='active' AND (c.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')",
        "hide": 2,
        "multi": true
      }
    ]
  }
}
```

**说明**:
- `bandwidth_line`: 用户可见，用于选择带宽线路
- `pop_machines`: 隐藏变量，存储POP机器的intra_ip列表
- `user_machines`: 隐藏变量，存储用户机器的intra_ip列表

#### 2. 更新Prometheus查询

**客户端POP指标** (使用 `exported_instance`):
```promql
# 旧查询 (错误)
pop_version{bandwidth_line_code=~"$bandwidth_line"}

# 新查询 (正确)
pop_version{exported_instance=~"$pop_machines"}
```

**客户端用户流量指标** (使用 `user_machine_ip`):
```promql
# 旧查询 (错误)
pop_traffic_tx_rate{bandwidth_line_code=~"$bandwidth_line"}

# 新查询 (正确)
pop_traffic_tx_rate{user_machine_ip=~"$user_machines"}
```

**服务端POP延迟指标** (使用 `pop_intra_ip`):
```promql
# 旧查询 (错误)
server_pop_latency{bandwidth_line_code=~"$bandwidth_line"}

# 新查询 (正确)
server_pop_latency{pop_intra_ip=~"$pop_machines"}
```

#### 3. 变量值格式处理

Grafana变量默认返回逗号分隔的值，需要使用正则匹配：

```
变量值: 10.1.0.1,10.1.0.2,10.1.0.3
Prometheus: exported_instance=~"10.1.0.1|10.1.0.2|10.1.0.3"
```

Grafana自动处理：`=~"$pop_machines"` → `=~"10.1.0.1|10.1.0.2|10.1.0.3"`

### 方案2: 在客户端/服务端代码中添加标签 (不推荐)

**原理**: 在暴露Prometheus指标时添加 `bandwidth_line_code` 标签

**优点**:
- 查询简单，直接使用 `bandwidth_line_code` 筛选
- 不需要额外的数据库查询

**缺点**:
- 需要修改客户端和服务端代码
- 配置变更时需要重启服务
- 标签基数增加（每个机器 × 每条线路）
- 违反Prometheus最佳实践（标签应该是指标的固有属性）

**不推荐原因**:
- 带宽线路是业务关系，不是指标固有属性
- 配置动态变化时需要重启服务更新标签
- 增加系统复杂度

### 方案3: 使用Prometheus Recording Rules (过度工程)

**原理**: 在Prometheus中创建预聚合规则，提前关联带宽线路

**示例**:
```yaml
groups:
  - name: bandwidth_line_enrichment
    interval: 30s
    rules:
      - record: pop_version:bandwidth_line
        expr: |
          pop_version * on(exported_instance) group_left(bandwidth_line_code)
          (mysql_bandwidth_lines_info)
```

**缺点**:
- 需要额外的MySQL exporter导出关系数据
- 配置复杂，维护成本高
- 实时性差（依赖recording rule计算周期）

## 推荐实现

### 优点

1. **逻辑清晰**: 筛选逻辑分离在变量定义中
2. **性能好**: 数据库查询简单，Prometheus查询高效
3. **灵活**: 不需要修改客户端/服务端代码
4. **实时**: 配置变更后立即生效（变量刷新）

### 注意事项

1. **MySQL数据源UID**: 确保配置文件中的 `mysql.uid` 与Grafana中一致
2. **变量刷新**: 设置 `refresh: 1` 确保变量自动刷新
3. **All选项处理**: SQL中使用 `OR '$bandwidth_line' = 'All'` 支持查看所有线路
4. **多值变量**: `pop_machines` 和 `user_machines` 设置为 `multi: true`

### SQL查询说明

#### POP机器查询
```sql
SELECT DISTINCT m.intra_ip 
FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) 
WHERE bl.is_active=1 
  AND bl.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND m.type='pop' 
  AND (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')
```

**解释**:
- 关联 `bandwidth_lines` 和 `machines` 表
- 获取A端和B端POP机器
- 筛选活跃线路和POP类型机器
- 支持All选项查看所有机器

#### 用户机器查询
```sql
SELECT DISTINCT m.intra_ip 
FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code 
WHERE c.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND c.status='active' 
  AND (c.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')
```

**解释**:
- 关联 `configs` 和 `machines` 表
- 获取配置中的用户机器
- 筛选活跃配置
- 支持All选项

## 面板更新清单

### 需要更新的客户端POP指标

| 面板 | 旧查询 | 新查询 |
|------|--------|--------|
| POP客户端软件版本号 | `pop_version{bandwidth_line_code=~"$bandwidth_line"}` | `pop_version{exported_instance=~"$pop_machines"}` |
| WireGuard peer连接数 | `pop_wireguard_peer_count{bandwidth_line_code=~"$bandwidth_line"}` | `pop_wireguard_peer_count{exported_instance=~"$pop_machines"}` |
| 与各peer节点连接状态 | `pop_wireguard_peer_status{bandwidth_line_code=~"$bandwidth_line"}` | `pop_wireguard_peer_status{exported_instance=~"$pop_machines"}` |
| POP客户端存活状态 | `pop_alive_status{bandwidth_line_code=~"$bandwidth_line"}` | `pop_alive_status{exported_instance=~"$pop_machines"}` |
| 域名解析服务状态 | `pop_dns_service_status{bandwidth_line_code=~"$bandwidth_line"}` | `pop_dns_service_status{exported_instance=~"$pop_machines"}` |
| 各带宽线路对端网络延迟 | `pop_wireguard_latency{bandwidth_line_code=~"$bandwidth_line"}` | `pop_wireguard_latency{exported_instance=~"$pop_machines"}` |

### 需要更新的客户端用户流量指标

| 面板 | 旧查询 | 新查询 |
|------|--------|--------|
| 各用户机器发送流量速率 | `pop_traffic_tx_rate{bandwidth_line_code=~"$bandwidth_line"}` | `pop_traffic_tx_rate{user_machine_ip=~"$user_machines"}` |
| 各用户机器接收流量速率 | `pop_traffic_rx_rate{bandwidth_line_code=~"$bandwidth_line"}` | `pop_traffic_rx_rate{user_machine_ip=~"$user_machines"}` |
| 每个用户机器的发送字节数 | `pop_wireguard_tx_bytes{bandwidth_line_code=~"$bandwidth_line"}` | `pop_wireguard_tx_bytes{user_machine_ip=~"$user_machines"}` |
| 每个用户机器的接收字节数 | `pop_wireguard_rx_bytes{bandwidth_line_code=~"$bandwidth_line"}` | `pop_wireguard_rx_bytes{user_machine_ip=~"$user_machines"}` |
| 用户上传限速触发 | `pop_rate_limit_hit{direction="upload",bandwidth_line_code=~"$bandwidth_line"}` | `pop_rate_limit_hit{direction="upload",user_machine_ip=~"$user_machines"}` |
| 用户下载限速触发 | `pop_rate_limit_hit{direction="download",bandwidth_line_code=~"$bandwidth_line"}` | `pop_rate_limit_hit{direction="download",user_machine_ip=~"$user_machines"}` |

### 需要更新的服务端指标

| 面板 | 旧查询 | 新查询 |
|------|--------|--------|
| 服务端到POP延迟 | `server_pop_latency{bandwidth_line_code=~"$bandwidth_line"}` | `server_pop_latency{pop_intra_ip=~"$pop_machines"}` |
| POP端与服务端通信状态 | `server_pop_communication_status{bandwidth_line_code=~"$bandwidth_line"}` | `server_pop_communication_status{pop_intra_ip=~"$pop_machines"}` |

### 不需要更新的面板（使用数据库查询）

- 服务端软件版本号
- 服务端健康状态
- 各用户订单个数
- 业务统计
- 各线路带宽总量
- 各线路带宽被购买使用情况

这些面板直接查询MySQL，已经使用 `bandwidth_line_code` 筛选，无需修改。

## 实施步骤

1. ✅ 更新 `business-base.json`，添加 `pop_machines` 和 `user_machines` 变量
2. ⏳ 更新所有客户端POP指标面板的查询
3. ⏳ 更新所有客户端用户流量指标面板的查询
4. ⏳ 更新所有服务端指标面板的查询
5. ⏳ 测试验证筛选功能
6. ⏳ 更新文档

## 测试验证

### 1. 验证变量级联

```sql
-- 测试POP机器查询
SELECT DISTINCT m.intra_ip 
FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) 
WHERE bl.is_active=1 
  AND bl.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND m.type='pop' 
  AND bl.bandwidth_line_code = 'LINE001';

-- 测试用户机器查询
SELECT DISTINCT m.intra_ip 
FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code 
WHERE c.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND c.status='active' 
  AND c.bandwidth_line_code = 'LINE001';
```

### 2. 验证Prometheus查询

```promql
# 测试POP版本查询
pop_version{exported_instance=~"10.1.0.1|10.1.0.2"}

# 测试用户流量查询
pop_traffic_tx_rate{user_machine_ip=~"10.1.1.1|10.1.1.2"}
```

### 3. 验证Grafana面板

1. 选择不同的带宽线路
2. 检查变量 `pop_machines` 和 `user_machines` 的值是否正确更新
3. 检查面板是否只显示对应线路的数据
4. 测试 "All" 选项是否显示所有数据

## 相关文档

- [MySQL数据源配置](MYSQL_DATASOURCE.md)
- [带宽线路筛选功能](BANDWIDTH_LINE_FILTER.md)
- [快速配置指南](QUICK_SETUP.md)
