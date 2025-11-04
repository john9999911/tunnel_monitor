# 按用户分类监控 Packet Rate 需求分析

## 需求概述

**目标**: 在客户端监控面板中，能够按用户维度查看每个用户的 Packet Rate（数据包速率），而不是只显示整个接口的聚合数据。

**使用场景**: 
- 管理员需要查看每个用户的流量使用情况
- 排查特定用户的网络问题
- 监控用户级别的流量异常

## 现状分析

### 当前实现方式

1. **指标收集方式**:
   - 从 `/sys/class/net/wg0/statistics/rx_packets` 和 `tx_packets` 读取
   - 这是整个接口的聚合统计，无法区分用户

2. **指标定义**:
   ```go
   rxPacketsGauge = prometheus.NewGaugeVec(
       prometheus.GaugeOpts{
           Name: "net_interface_rx_packets",
           Help: "RX packets as reported by sysfs (monotonic)",
       },
       []string{"iface"},  // 只有接口标签，没有用户标签
   )
   ```

3. **Grafana查询**:
   ```promql
   rate(net_interface_rx_packets{iface="wg0"}[1m])
   ```

### 系统现有能力

1. **用户标识**:
   - 每个用户有唯一的 `user_intra_ip`（内网IP）
   - 配置中包含 `UserIntraIP` 字段

2. **流量标记机制**:
   - 系统使用 iptables 的 MARK 功能标记流量
   - 每个用户的流量规则有对应的 `markId`
   - `TcMarkIdPool` 保存了 markId 到 IP 的映射关系

3. **iptables规则统计**:
   - iptables 规则包含 `pkts` 字段，记录匹配的数据包数量
   - 可以通过 `iptables -t mangle -L -n -v` 获取统计

## 实现方案分析

### 方案一：通过 iptables 规则统计（推荐）

**原理**: 
- 从 iptables 规则的统计信息中提取每个 mark 的数据包计数
- 通过 markId 映射到 user_intra_ip
- 通过配置映射 user_intra_ip 到 user_code

**优点**:
- ✅ 利用现有 iptables 规则，无需额外开销
- ✅ 统计准确，直接反映实际流量
- ✅ 实现相对简单

**缺点**:
- ⚠️ 需要解析 iptables 输出（文本解析）
- ⚠️ 需要维护 markId → user_intra_ip → user_code 的映射关系

**实现步骤**:
1. 在指标收集时，解析 iptables 规则统计
2. 提取每个 mark 的 pkts 计数
3. 通过映射关系获取 user_code
4. 更新 Prometheus 指标，添加 `user_code` 标签

**代码修改位置**:
- `tunnel_client/internal/monitor/collectors.go` 的 `updateTCMetrics()` 或新增函数
- 需要访问配置管理器获取 user_code 映射

### 方案二：通过 TC 规则统计

**原理**:
- 从 TC filter 的统计信息中提取数据包计数
- TC 规则也使用 markId 进行分类

**优点**:
- ✅ 统计准确
- ✅ 与流量控制逻辑一致

**缺点**:
- ⚠️ TC 统计获取相对复杂
- ⚠️ 需要解析 `tc -s filter show` 输出

### 方案三：通过 eBPF 直接统计（不推荐）

**原理**:
- 使用 eBPF 在内核层面统计每个用户的包数

**优点**:
- ✅ 性能最好
- ✅ 统计最准确

**缺点**:
- ❌ 实现复杂度高
- ❌ 需要 eBPF 环境支持
- ❌ 维护成本高

## 推荐方案：方案一（iptables统计）

### 详细实现设计

#### 1. 指标定义修改

```go
// 新增按用户的packet统计指标
userRxPacketsGauge = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "net_user_rx_packets",
        Help: "RX packets per user (monotonic counter)",
    },
    []string{"user_code", "user_intra_ip"},  // 添加用户标签
)

userTxPacketsGauge = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "net_user_tx_packets",
        Help: "TX packets per user (monotonic counter)",
    },
    []string{"user_code", "user_intra_ip"},
)
```

#### 2. 数据收集流程

```
1. 读取 iptables 规则统计
   ↓
2. 解析每个 mark 的 pkts 计数
   ↓
3. 通过 markId 查找对应的 IP (sips/dips)
   ↓
4. 通过 IP 查找 user_code（需要配置管理器支持）
   ↓
5. 更新 Prometheus 指标
```

#### 3. 需要解决的关键问题

**问题1: markId → user_code 映射**
- **现状**: 只有 markId → IP 映射（在 TcMarkIdPool 中）
- **需求**: 需要 IP → user_code 映射
- **解决方案**: 
  - 从配置管理器中获取 Config 列表
  - 建立 `user_intra_ip → user_code` 的映射表
  - 或者修改配置同步逻辑，保存 user_code 信息

**问题2: 区分上传和下载**
- **现状**: iptables 规则在 POSTROUTING 和 PREROUTING 链中
- **解决方案**:
  - POSTROUTING 链的统计 = 上传（TX）
  - PREROUTING 链的统计 = 下载（RX）

**问题3: 多个 mark 对应同一用户**
- **现状**: 一个用户可能有多个 mark（不同目标IP）
- **解决方案**: 
  - 聚合同一用户的所有 mark 的统计
  - 或者分别显示每个目标IP的统计

#### 4. Grafana 面板配置

**全局变量**:
```json
{
  "name": "User",
  "type": "query",
  "query": "label_values(net_user_rx_packets, user_code)",
  "multi": true,
  "includeAll": true
}
```

**查询表达式**:
```promql
# 按用户查询 RX Packet Rate
rate(net_user_rx_packets{user_code=~"$User"}[1m])

# 按用户查询 TX Packet Rate  
rate(net_user_tx_packets{user_code=~"$User"}[1m])
```

## 实施建议

### 阶段一：基础实现（推荐先做）

1. **修改指标定义**，添加 user_code 标签
2. **实现 iptables 统计解析**，获取每个 mark 的包数
3. **建立 markId → user_intra_ip 映射**（利用现有 TcMarkIdPool）
4. **实现基础的用户级统计**（先用 user_intra_ip 作为标识）

**说明**: 这个阶段可以实现按 IP 分类的监控，虽然不如 user_code 直观，但可以先验证方案可行性。

### 阶段二：完善实现

1. **获取 user_code 映射**
   - 从配置管理器获取 Config 列表
   - 建立 user_intra_ip → user_code 映射
2. **添加 Grafana 全局变量**
3. **优化面板显示**，按用户分组

### 阶段三：优化（可选）

1. **缓存映射关系**，减少查询开销
2. **处理动态配置变化**，及时更新映射
3. **添加用户名称显示**（如果配置中有）

## 技术挑战

1. **配置获取**: 需要从配置管理器获取 user_code 信息
2. **文本解析**: iptables 输出格式可能变化，需要健壮的解析
3. **性能**: 定期解析 iptables 输出可能有一定开销
4. **准确性**: iptables 规则删除后，统计可能不准确

## 风险评估

- **低风险**: 方案一实现相对简单，风险较低
- **中等风险**: 需要确保映射关系准确，否则统计会出错
- **建议**: 可以先实现基础版本，逐步完善

## 总结

**可行性**: ✅ 高
**复杂度**: 中等
**推荐方案**: 方案一（通过 iptables 统计）
**实施优先级**: 中等（取决于实际需求紧迫性）

建议先实现基础版本（按 IP 分类），验证可行性后再添加 user_code 映射。

