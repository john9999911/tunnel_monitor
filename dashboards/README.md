# Dashboard 模板文件说明

## 文件结构

模板文件已经拆分为更易管理的结构：

```
dashboards/
├── client-template.json          # 完整客户端模板（保留作为备份）
├── client-base.json              # 客户端基础模板（不含panels）
├── server-template.json          # 完整服务端模板（保留作为备份）
├── server-base.json              # 服务端基础模板（不含panels）
├── database-template.json        # 完整数据库模板（保留作为备份）
├── database-base.json            # 数据库基础模板（不含panels）
└── panels/                       # Panels目录
    ├── client/                   # 客户端panels（17个文件）
    │   ├── WireGuard_Status.json
    │   ├── Active_Connections.json
    │   ├── Bandwidth_Usage.json
    │   └── ...
    ├── server/                   # 服务端panels（8个文件）
    │   ├── HTTP请求率.json
    │   ├── HTTP响应时间.json
    │   └── ...
    └── database/                 # 数据库panels（6个文件）
        ├── 数据库连接数.json
        ├── 数据库查询耗时.json
        └── ...
```

## 拆分效果

### 原始文件大小
- `client-template.json`: 1545行, 35KB
- `server-template.json`: 612行, 14KB
- `database-template.json`: 425行, 9.9KB

### 拆分后
- **基础模板**：每个约20-30行，只包含dashboard元数据
- **Panels文件**：每个panel独立文件，易于管理和修改
- **总计**：31个panel文件，每个文件约50-300行

## 使用方式

### 自动加载

代码会自动检测并使用拆分后的模板：

1. **优先使用拆分模板**：如果存在 `*-base.json` 和对应的 `panels/` 目录，会自动组合使用
2. **回退到完整模板**：如果拆分模板不存在，会自动使用原始的完整模板文件

### 手动拆分模板

如果需要重新拆分模板文件，可以使用：

```bash
go run cmd/split-templates.go
```

这会：
- 读取完整的模板文件
- 提取所有panels到独立文件
- 创建基础模板文件（不含panels）

### 修改Panel

现在可以直接编辑单个panel文件，无需修改巨大的JSON文件：

```bash
# 编辑客户端某个panel
vim dashboards/panels/client/WireGuard_Status.json

# 编辑服务端某个panel
vim dashboards/panels/server/HTTP请求率.json
```

## 优势

1. **易于维护**：每个panel独立文件，修改不会影响其他panel
2. **版本控制友好**：Git diff可以精确显示哪个panel被修改
3. **代码组织清晰**：基础配置和面板分离，结构更清晰
4. **向后兼容**：仍然支持完整模板文件，拆分是可选的

## 模板加载逻辑

代码加载模板时的优先级：

```
LoadClientTemplate() / LoadServerTemplate() / LoadDatabaseTemplate()
    ↓
1. 检查是否存在拆分模板（*-base.json + panels/目录）
    ↓ 是 → 使用 TemplateManager 组合加载
    ↓ 否
2. 使用完整模板文件（*-template.json）
```

## 注意事项

1. **保持一致性**：如果修改了panel文件，确保基础模板和panel文件同步
2. **文件名**：Panel文件名基于panel标题自动生成，特殊字符会被替换为下划线
3. **顺序**：Panels按文件名排序加载，如需特定顺序，请使用数字前缀（如 `01_xxx.json`）
4. **备份**：原始的完整模板文件已保留，可作为备份参考
