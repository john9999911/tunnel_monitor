# 带宽线路筛选功能

## 概述

业务监控面板支持按带宽线路筛选监控数据。选择不同的带宽线路后，面板会自动显示该线路相关的所有指标数据。

## 实现原理

### 筛选逻辑

带宽线路筛选采用**变量级联**机制：

```
用户选择带宽线路 ($bandwidth_line)
    ↓
数据库查询相关机器 (MySQL)
    ├─ POP机器列表 ($pop_machines)
    └─ 用户机器列表 ($user_machines)
    ↓
Prometheus指标筛选
    ├─ 使用 exported_instance 匹配 POP 机器
    └─ 使用 user_machine_ip 匹配用户机器
```

### 为什么不直接使用 bandwidth_line_code？

Prometheus指标不包含 `bandwidth_line_code` 标签，因为：
1. **带宽线路是业务关系**，不是指标的固有属性
2. **配置动态变化**，如果在指标中添加标签需要重启服务
3. **标签基数问题**，会导致指标数量爆炸（机器数 × 线路数）

正确的做法是：在Grafana中通过MySQL查询获取机器列表，然后用机器标识筛选Prometheus指标。

## 全局变量定义

### 1. bandwidth_line (用户可见)

**用途**: 让用户选择要查看的带宽线路

**数据源**: MySQL

**查询**:
```sql
SELECT DISTINCT bandwidth_line_code 
FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL 
ORDER BY bandwidth_line_code
```

**配置**:
- `includeAll: true` - 支持选择"All"查看所有线路
- `multi: false` - 单选模式
- `refresh: 1` - 自动刷新

### 2. pop_machines (隐藏变量)

**用途**: 存储选定带宽线路的POP机器内网IP列表

**数据源**: MySQL

**查询**:
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

**配置**:
- `hide: 2` - 完全隐藏
- `multi: true` - 支持多值
- `refresh: 1` - 级联刷新

### 3. user_machines (隐藏变量)

**用途**: 存储选定带宽线路的用户机器内网IP列表

**数据源**: MySQL

**查询**:
```sql
SELECT DISTINCT m.intra_ip 
FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code 
WHERE c.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND c.status='active' 
  AND (c.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')
```

**配置**:
- `hide: 2` - 完全隐藏
- `multi: true` - 支持多值
- `refresh: 1` - 级联刷新

## Prometheus查询示例

### 客户端POP指标

使用 `exported_instance` 标签匹配 `$pop_machines`：

```promql
# POP客户端版本
pop_version{exported_instance=~"$pop_machines"}

# WireGuard peer连接数
pop_wireguard_peer_count{exported_instance=~"$pop_machines"}

# POP存活状态
pop_alive_status{exported_instance=~"$pop_machines"}

# 域名解析服务状态
pop_dns_service_status{exported_instance=~"$pop_machines"}

# 与peer节点连接状态
pop_wireguard_peer_status{exported_instance=~"$pop_machines"}

# 对端网络延迟
pop_wireguard_latency{exported_instance=~"$pop_machines"}
```

### 客户端用户流量指标

使用 `user_machine_ip` 标签匹配 `$user_machines`：

```promql
# 用户发送流量速率 (转换为 Kbps)
pop_traffic_tx_rate{user_machine_ip=~"$user_machines"} / 1000

# 用户接收流量速率 (转换为 Kbps)
pop_traffic_rx_rate{user_machine_ip=~"$user_machines"} / 1000

# 用户发送字节数
pop_wireguard_tx_bytes{user_machine_ip=~"$user_machines"}

# 用户接收字节数
pop_wireguard_rx_bytes{user_machine_ip=~"$user_machines"}

# 用户上传限速触发
pop_rate_limit_hit{direction="upload",user_machine_ip=~"$user_machines"}

# 用户下载限速触发
pop_rate_limit_hit{direction="download",user_machine_ip=~"$user_machines"}
```

### 服务端指标

使用 `pop_intra_ip` 标签匹配 `$pop_machines`：

```promql
# 服务端到POP延迟
server_pop_latency{pop_intra_ip=~"$pop_machines"}

# POP与服务端通信状态
server_pop_communication_status{pop_intra_ip=~"$pop_machines"}
```

## 使用说明

### 查看特定线路

1. 打开 IPTunnel 业务监控面板
2. 点击顶部的"带宽线路"下拉框
3. 选择要查看的线路（如：`LINE001`）
4. 面板自动刷新，只显示该线路的数据

### 查看所有线路

1. 点击"带宽线路"下拉框
2. 选择"All"选项
3. 面板显示所有线路的汇总数据

## 技术细节

### 变量值格式

Grafana变量自动处理多值转换：

```
数据库返回: ['10.1.0.1', '10.1.0.2', '10.1.0.3']
           ↓
变量值:     10.1.0.1,10.1.0.2,10.1.0.3
           ↓
Prometheus: exported_instance=~"10.1.0.1|10.1.0.2|10.1.0.3"
```

### All 选项处理

SQL查询中使用条件判断：

```sql
WHERE (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')
```

- 选择特定线路时：`bl.bandwidth_line_code = 'LINE001'`
- 选择All时：`'All' = 'All'` 永远为真，返回所有机器

### 级联刷新

变量设置 `refresh: 1` (on dashboard load)，确保：
1. 页面加载时自动刷新
2. `$bandwidth_line` 变化时，`$pop_machines` 和 `$user_machines` 自动更新
3. 变量更新后，Prometheus查询自动重新执行

## 故障排查

### 问题1: 选择线路后无数据显示

**可能原因**:
- 数据库中该线路没有关联的机器
- 机器状态不正确（deleted_at 不为 NULL）
- Prometheus中没有对应机器的数据

**排查步骤**:
```sql
-- 检查线路配置
SELECT * FROM bandwidth_lines 
WHERE bandwidth_line_code = 'LINE001' AND deleted_at IS NULL;

-- 检查POP机器
SELECT m.* FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code)
WHERE bl.bandwidth_line_code = 'LINE001' AND m.deleted_at IS NULL;

-- 检查用户机器
SELECT m.* FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code
WHERE c.bandwidth_line_code = 'LINE001' AND c.deleted_at IS NULL;
```

### 问题2: 隐藏变量值为空

**原因**: 数据库查询没有返回结果

**解决**:
1. 检查数据库连接是否正常
2. 验证SQL查询是否正确
3. 确认数据库中有相应的数据

### 问题3: "All"选项不工作

**原因**: SQL条件判断逻辑错误

**检查**:
```sql
-- 测试All条件
SELECT * FROM bandwidth_lines 
WHERE (bandwidth_line_code = 'LINE001' OR 'All' = 'All');

-- 应该返回所有记录
```

## 性能优化

### 数据库查询优化

1. **索引**: 确保相关字段有索引
   ```sql
   -- bandwidth_lines表
   KEY idx_bandwidth_line_code (bandwidth_line_code)
   KEY idx_machine_a_code (machine_a_code)
   KEY idx_machine_b_code (machine_b_code)
   
   -- configs表
   KEY idx_bandwidth_line_code (bandwidth_line_code)
   KEY idx_user_machine_code (user_machine_code)
   
   -- machines表
   KEY idx_machine_code (machine_code)
   KEY idx_type (type)
   ```

2. **查询缓存**: Grafana会缓存变量查询结果

3. **避免全表扫描**: 使用 `deleted_at IS NULL` 而不是 `deleted_at = NULL`

### Prometheus查询优化

1. **使用正则匹配**: `=~` 比多个 `=` 更高效
2. **减少标签基数**: 只使用必要的标签进行过滤
3. **合理的查询周期**: 避免查询过长的时间范围

## 相关文档

- [筛选策略详解](FILTERING_STRATEGY.md)
- [MySQL数据源配置](MYSQL_DATASOURCE.md)
- [快速配置指南](QUICK_SETUP.md)
