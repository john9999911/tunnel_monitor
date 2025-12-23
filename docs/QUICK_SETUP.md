# 监控面板快速配置指南

## 问题：数据库查询指标无法使用

### 原因分析

监控面板中有以下指标需要查询MySQL数据库：
1. **全局变量**: 带宽线路下拉框 (`$bandwidth_line`)
2. **服务端面板**: 
   - 各用户订单个数
   - 业务统计（POP数量、用户数、订单数）
   - 各线路带宽总量
   - 各线路带宽被购买使用情况

这些查询使用Grafana的MySQL数据源，但面板创建时需要知道MySQL数据源的UID才能正确引用。

### 解决方案

通过配置文件指定MySQL数据源的UID，在创建面板时自动替换模板中的 `{{MYSQL_UID}}` 占位符。

## 配置步骤

### 步骤1: 在Grafana中创建MySQL数据源

1. **登录Grafana**
   ```
   URL: http://localhost:3000
   用户名: admin
   密码: admin
   ```

2. **创建数据源**
   - 点击左侧菜单 ⚙️ (Configuration)
   - 选择 "Data sources"
   - 点击 "Add data source"
   - 选择 "MySQL"

3. **配置连接**
   ```
   Name: iptunnel-mysql (或任意名称)
   Host: localhost:3306
   Database: iptunnel
   User: root
   Password: your_password
   ```

4. **保存并测试**
   - 点击 "Save & test"
   - 确保显示 "Database Connection OK"

5. **获取数据源UID**
   - 保存后，查看浏览器地址栏
   - URL格式：`http://localhost:3000/datasources/edit/<UID>`
   - 例如：`http://localhost:3000/datasources/edit/mysql-datasource`
   - 记录这个UID（如：`mysql-datasource`）

### 步骤2: 配置 tunnel_monitor

编辑 `config.yaml` 文件：

```yaml
# MySQL数据源配置（用于面板查询）
mysql:
  host: "localhost"
  port: 3306
  database: "iptunnel"
  username: "root"
  password: "your_password"
  uid: "mysql-datasource"  # ⚠️ 这里填写步骤1中获取的UID
```

**关键点**:
- `uid` 必须与Grafana中MySQL数据源的实际UID完全一致
- 其他字段（host, port等）用于文档说明，实际连接配置在Grafana数据源中

### 步骤3: 创建监控面板

```bash
cd /home/ubuntu/src/tunnel_monitor
./tunnel-monitor dashboard create
```

程序会自动：
1. 加载面板模板
2. 将所有 `{{MYSQL_UID}}` 替换为配置文件中的 `mysql.uid`
3. 导入到Grafana

### 步骤4: 验证

1. **打开Grafana面板**
   ```
   http://localhost:3000/d/iptunnel-business
   ```

2. **检查带宽线路下拉框**
   - 顶部应该有"带宽线路"下拉框
   - 点击下拉框应该能看到线路列表
   - 如果为空，检查数据库是否有数据：
     ```sql
     SELECT * FROM bandwidth_lines WHERE is_active=1;
     ```

3. **检查数据库查询面板**
   - "各用户订单个数"面板应该显示数据
   - "业务统计"面板应该显示POP数量、用户数等
   - 如果显示错误，检查MySQL数据源配置

## 常见问题

### Q1: 带宽线路下拉框显示 "No data"

**原因**: 数据库中没有符合条件的带宽线路数据

**解决**:
```sql
-- 检查数据
SELECT bandwidth_line_code FROM bandwidth_lines 
WHERE is_active=1 AND deleted_at IS NULL;

-- 如果没有数据，插入测试数据
INSERT INTO bandwidth_lines (bandwidth_line_code, is_active) 
VALUES ('LINE001', 1);
```

### Q2: 面板显示 "Data source not found"

**原因**: `config.yaml` 中的 `mysql.uid` 与Grafana中的数据源UID不匹配

**解决**:
1. 在Grafana中检查MySQL数据源的UID
2. 更新 `config.yaml` 中的 `mysql.uid`
3. 重新创建面板：`./tunnel-monitor dashboard create`

### Q3: MySQL连接测试失败

**原因**: 数据库连接信息错误或权限不足

**解决**:
```bash
# 测试数据库连接
mysql -h localhost -u root -p iptunnel

# 检查用户权限
GRANT ALL PRIVILEGES ON iptunnel.* TO 'root'@'localhost';
FLUSH PRIVILEGES;
```

### Q4: 修改配置后面板仍然不工作

**原因**: 需要重新创建面板才能应用新配置

**解决**:
```bash
# 重新创建面板
./tunnel-monitor dashboard create

# 或者在Grafana中手动删除旧面板后重新创建
```

## 技术细节

### 占位符替换机制

面板模板中使用 `{{MYSQL_UID}}` 作为占位符：

```json
{
    "datasource": {
        "type": "mysql",
        "uid": "{{MYSQL_UID}}"
    }
}
```

创建面板时，`FixDatasource()` 函数会递归遍历面板JSON，将所有 `{{MYSQL_UID}}` 替换为配置文件中的实际UID。

### 代码位置

- 配置定义: `internal/config/config.go`
- 数据源修复: `internal/dashboard/builder.go` - `FixDatasource()`
- 面板创建: `internal/dashboard/dashboard.go` - `CreateBusinessDashboard()`

## 完整工作流程示例

```bash
# 1. 确保MySQL数据库运行
sudo systemctl start mysql

# 2. 创建数据库和表（如果还没有）
mysql -u root -p < tunnel_server/desc/sql/schema.sql

# 3. 启动Grafana
sudo ./tunnel-monitor start

# 4. 在Grafana中配置MySQL数据源
# - 浏览器访问 http://localhost:3000
# - 创建MySQL数据源
# - 记录数据源UID

# 5. 配置tunnel_monitor
vim config.yaml
# 更新 mysql.uid 字段

# 6. 创建监控面板
./tunnel-monitor dashboard create

# 7. 验证
# - 打开面板 http://localhost:3000/d/iptunnel-business
# - 检查带宽线路下拉框
# - 检查数据库查询面板
```

## 相关文档

- [MySQL数据源详细配置](MYSQL_DATASOURCE.md)
- [带宽线路筛选功能](BANDWIDTH_LINE_FILTER.md)
- [面板模板结构](../dashboards/README.md)
