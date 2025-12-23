# Grafana变量配置验证清单

## ✅ 配置验证

运行以下脚本验证配置：

### 1. 验证JSON配置
```bash
cd /home/ubuntu/src/tunnel_monitor
./scripts/verify_variables.sh
```

**期望结果**：
- ✓ 发现 allValue 字段
- ✓ pop_machines 配置正确
- ✓ user_machines 配置正确

### 2. 验证变量SQL查询
```bash
cd /home/ubuntu/src/tunnel_monitor
./scripts/test_variable_queries.sh
```

**期望结果**：
- ✓ 带宽线路查询：返回活跃线路
- ✓ POP机器查询（All和特定线路）：返回机器IP
- ✓ 用户机器查询（All和特定线路）：返回机器IP

### 3. 验证MySQL面板查询
```bash
cd /home/ubuntu/src/tunnel_monitor
./scripts/test_mysql_panels.sh
```

**期望结果**：
- ✓ 各线路带宽总量 Query A：All和特定线路都成功
- ✓ 各线路带宽总量 Query B：All和特定线路都成功
- ✓ 带宽被购买使用情况：All和特定线路都成功

## 📋 配置要点检查表

### 主变量（bandwidth_line）
- [ ] `includeAll: true`
- [ ] `allValue: "All"` （注意：是字符串"All"，不是正则表达式）
- [ ] `multi: false`（单选）
- [ ] `refresh: 1`
- [ ] `datasource.type: "mysql"`
- [ ] `datasource.uid` 匹配 config.yaml

### 派生变量（pop_machines）
- [ ] `includeAll: true`
- [ ] `allValue: ".*"` （正则表达式，用于Prometheus匹配）
- [ ] `multi: true`（多选）
- [ ] `hide: 2`（隐藏）
- [ ] `refresh: 1`
- [ ] 查询包含 `(bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')`

### 派生变量（user_machines）
- [ ] `includeAll: true`
- [ ] `allValue: ".*"`
- [ ] `multi: true`
- [ ] `hide: 2`
- [ ] `refresh: 1`
- [ ] 查询包含 `(c.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')`

### MySQL面板查询
- [ ] 使用条件判断：`('$bandwidth_line' = 'All' OR bandwidth_line_code = '$bandwidth_line')`
- [ ] **不要**使用 `REGEXP '$bandwidth_line'`
- [ ] 所有面板都已更新：
  - [ ] 各线路带宽总量（2个查询）
  - [ ] 各线路带宽被购买使用情况（1个查询）

## 🔍 常见问题排查

### 问题1：变量显示为空

**检查步骤**：
1. 验证数据库连接：
   ```bash
   mysql -u root tunnel -e "SELECT COUNT(*) FROM bandwidth_lines"
   ```
2. 检查MySQL数据源UID：
   ```bash
   curl -s -u admin:admin http://localhost:3000/api/datasources | grep mysql
   ```
3. 确认UID匹配 config.yaml 中的设置

### 问题2：SQL语法错误

**症状**：Error 1064, "near 'All GROUP BY"

**原因**：使用了 `REGEXP '$bandwidth_line'`，当 `$bandwidth_line = 'All'` 时导致语法错误

**解决**：改用条件判断 `('$bandwidth_line' = 'All' OR bandwidth_line_code = '$bandwidth_line')`

### 问题3：选择All后面板无数据

**检查**：
1. 主变量是否有 `allValue: "All"`
2. 派生变量是否有 `allValue: ".*"`
3. SQL条件是否正确：`OR '$bandwidth_line' = 'All'`

### 问题4：切换线路后面板不更新

**检查**：
1. 所有变量的 `refresh: 1` 是否设置
2. 派生变量的查询是否包含 `$bandwidth_line`
3. 浏览器是否需要刷新

## 📊 手动验证步骤

### 1. 在Grafana中检查变量

1. 打开dashboard：http://localhost:3000/d/iptunnel-business
2. 点击⚙️ → Variables
3. 检查每个变量的配置：
   - bandwidth_line：有"All"选项
   - pop_machines：Preview显示IP列表
   - user_machines：Preview显示IP列表

### 2. 测试筛选功能

1. 选择"带宽线路"为"All"：
   - 观察所有面板是否显示数据
   - 检查是否显示所有机器的指标

2. 选择特定带宽线路：
   - 观察面板是否只显示该线路的数据
   - 确认POP机器和用户机器都被正确筛选

### 3. 检查Inspector

1. 打开任意面板
2. 点击标题 → Inspect → Query
3. 查看实际执行的查询：
   - 变量是否被正确替换
   - SQL语法是否正确

## 🚀 重新部署

如果修改了配置文件，需要重新部署dashboard：

```bash
cd /home/ubuntu/src/tunnel_monitor
go run main.go dashboard create
```

然后在浏览器中刷新页面（Ctrl+Shift+R 强制刷新）。

## 📝 配置文件位置

- **模板文件**：
  - `dashboards/business-base.json`
  - `dashboards/business-template.json`
  
- **面板文件**：
  - `dashboards/panels/client/*.json` （12个）
  - `dashboards/panels/server/*.json` （8个）

- **测试脚本**：
  - `scripts/verify_variables.sh`
  - `scripts/test_variable_queries.sh`
  - `scripts/test_mysql_panels.sh`

## 🎯 成功标准

✅ **所有以下条件都满足**：
- [ ] verify_variables.sh 通过
- [ ] test_variable_queries.sh 通过
- [ ] test_mysql_panels.sh 通过
- [ ] Grafana中变量有预览值
- [ ] 选择"All"显示所有数据
- [ ] 选择特定线路只显示该线路数据
- [ ] MySQL面板不报SQL错误
- [ ] 所有20个面板都正常显示

---

**最后更新**：2025-12-23  
**验证人**：GitHub Copilot
