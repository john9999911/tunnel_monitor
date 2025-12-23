# MySQL 数据源配置说明

## 概述

监控面板中有部分指标需要查询MySQL数据库，包括：
- **全局变量**: 带宽线路下拉框 (`$bandwidth_line`)
- **服务端面板**: 各用户订单个数、业务统计、各线路带宽总量、各线路带宽被购买使用情况

这些查询依赖Grafana中配置的MySQL数据源。

## 配置步骤

### 1. 在Grafana中创建MySQL数据源

1. 登录Grafana (默认: http://localhost:3000)
2. 进入 **Configuration** → **Data sources**
3. 点击 **Add data source**
4. 选择 **MySQL**
5. 配置连接信息:
   - **Name**: 任意名称（如：`iptunnel-mysql`）
   - **Host**: `localhost:3306`（根据实际情况修改）
   - **Database**: `iptunnel`
   - **User**: 数据库用户名
   - **Password**: 数据库密码
6. 点击 **Save & test** 测试连接
7. **记录数据源的UID**（在保存后的页面URL中可以看到，如：`mysql-datasource`）

### 2. 配置 tunnel_monitor

编辑 `config.yaml` 文件，添加MySQL配置：

```yaml
# MySQL数据源配置（用于面板查询）
mysql:
  host: "localhost"
  port: 3306
  database: "iptunnel"
  username: "root"
  password: "your_password"
  uid: "mysql-datasource"  # 这里填写Grafana中MySQL数据源的实际UID
```

**重要**: `uid` 字段必须与Grafana中MySQL数据源的UID完全一致！

### 3. 创建面板

运行命令创建监控面板：

```bash
./tunnel-monitor dashboard create
```

程序会自动将面板模板中的 `{{MYSQL_UID}}` 占位符替换为配置文件中的 `mysql.uid`。

## 验证

1. 在Grafana中打开 **IPTunnel 业务监控** 面板
2. 查看顶部的 **带宽线路** 下拉框是否正常加载数据
3. 查看服务端相关的数据库查询面板是否正常显示数据

## 故障排查

### 问题1: 带宽线路下拉框为空

**原因**: 
- MySQL数据源配置错误
- 数据库中 `bandwidth_lines` 表为空或没有 `is_active=1` 的记录

**解决**:
```sql
-- 检查数据
SELECT bandwidth_line_code FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL;
```

### 问题2: 数据库查询面板显示 "Data source not found"

**原因**: 配置文件中的 `mysql.uid` 与Grafana中实际的数据源UID不一致

**解决**:
1. 在Grafana中查看MySQL数据源的UID（URL: `/datasources/edit/<UID>`）
2. 更新 `config.yaml` 中的 `mysql.uid`
3. 重新创建面板：`./tunnel-monitor dashboard create`

### 问题3: 查询返回空数据

**原因**: 
- 数据库表结构变化
- 查询SQL与实际表结构不匹配
- 没有相关数据

**解决**:
检查相关表的数据：
```sql
-- 检查订单数据
SELECT user_code, COUNT(*) FROM orders GROUP BY user_code;

-- 检查带宽配置
SELECT * FROM configs WHERE bandwidth_line_code REGEXP '^YOUR_LINE_CODE';

-- 检查POP机器
SELECT COUNT(*) FROM machines WHERE machine_type='pop';
```

## 面板中使用MySQL的查询

### 全局变量 - 带宽线路

```sql
SELECT DISTINCT bandwidth_line_code 
FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL 
ORDER BY bandwidth_line_code
```

### 各用户订单个数

```sql
SELECT user_code, COUNT(*) as order_count
FROM orders
WHERE bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All'
GROUP BY user_code
```

### 业务统计

```sql
-- POP机器数量
SELECT COUNT(*) as pop_count FROM machines WHERE machine_type='pop';

-- 用户数量
SELECT COUNT(DISTINCT user_code) as user_count FROM users;

-- 订单数量
SELECT COUNT(*) as order_count FROM orders;
```

### 各线路带宽总量

```sql
SELECT bandwidth_line_code, total_bandwidth
FROM bandwidth_lines
WHERE bandwidth_line_code REGEXP '$bandwidth_line' OR '$bandwidth_line' = 'All'
```

### 各线路带宽被购买使用情况

```sql
SELECT bandwidth_line_code, SUM(bandwidth) as used_bandwidth
FROM configs
WHERE bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All'
GROUP BY bandwidth_line_code
```

## 注意事项

1. **数据源UID必须一致**: 配置文件中的 `mysql.uid` 必须与Grafana中MySQL数据源的实际UID完全一致
2. **重新创建面板**: 修改配置后需要重新运行 `dashboard create` 命令
3. **数据库连接**: 确保tunnel_monitor运行环境能够访问MySQL数据库
4. **表结构依赖**: 查询依赖特定的表结构，如果数据库schema变化需要更新面板查询

## 相关文档

- [带宽线路筛选功能](BANDWIDTH_LINE_FILTER.md)
- [面板模板说明](../dashboards/README.md)
