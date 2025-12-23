#!/bin/bash

# 验证使用%和LIKE语句的解决方案
# 此方案避免了Grafana变量替换时引号处理不一致的问题

set -e

echo "======================================"
echo "Grafana全局变量 - LIKE方案验证"
echo "======================================"
echo ""

# 测试1: MySQL LIKE '%' - 匹配所有
echo "✅ 测试1: LIKE '%' 匹配所有带宽线路"
echo "--------------------------------------"
RESULT=$(mysql -u root tunnel -N -e "SELECT COUNT(*) FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL AND bandwidth_line_code LIKE '%';")
echo "查询结果: $RESULT 条记录"
if [ "$RESULT" -ge 2 ]; then
    echo "✅ 通过：返回所有带宽线路"
else
    echo "❌ 失败：期望至少2条，实际$RESULT条"
    exit 1
fi
echo ""

# 测试2: MySQL LIKE 特定线路 - 精确匹配
echo "✅ 测试2: LIKE '特定线路' 精确匹配"
echo "--------------------------------------"
RESULT=$(mysql -u root tunnel -N -e "SELECT COUNT(*) FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL AND bandwidth_line_code LIKE 'BANDWIDTH_LINE_1996819610904563712';")
echo "查询结果: $RESULT 条记录"
if [ "$RESULT" -eq 1 ]; then
    echo "✅ 通过：精确匹配特定线路"
else
    echo "❌ 失败：期望1条，实际$RESULT条"
    exit 1
fi
echo ""

# 测试3: 派生变量 pop_machines with %
echo "✅ 测试3: 派生变量 pop_machines (LIKE '%')"
echo "--------------------------------------"
RESULT=$(mysql -u root tunnel -N -e "SELECT COUNT(DISTINCT m.intra_ip) FROM bandwidth_lines bl JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) WHERE bl.is_active=1 AND bl.deleted_at IS NULL AND m.deleted_at IS NULL AND m.type='pop' AND bl.bandwidth_line_code LIKE '%';")
echo "查询结果: $RESULT 个POP机器"
if [ "$RESULT" -ge 2 ]; then
    echo "✅ 通过：返回所有POP机器"
else
    echo "❌ 失败：期望至少2个，实际$RESULT个"
    exit 1
fi
echo ""

# 测试4: 派生变量 user_machines with %
echo "✅ 测试4: 派生变量 user_machines (LIKE '%')"
echo "--------------------------------------"
RESULT=$(mysql -u root tunnel -N -e "SELECT COUNT(DISTINCT m.intra_ip) FROM configs c JOIN machines m ON c.user_machine_code = m.machine_code WHERE c.deleted_at IS NULL AND m.deleted_at IS NULL AND c.status='active' AND c.bandwidth_line_code LIKE '%';")
echo "查询结果: $RESULT 个用户机器"
if [ "$RESULT" -ge 2 ]; then
    echo "✅ 通过：返回所有用户机器"
else
    echo "❌ 失败：期望至少2个，实际$RESULT个"
    exit 1
fi
echo ""

# 测试5: GROUP BY查询 with %
echo "✅ 测试5: GROUP BY查询 (LIKE '%')"
echo "--------------------------------------"
RESULT=$(mysql -u root tunnel -N -e "SELECT COUNT(*) FROM (SELECT bandwidth_line_code, user_code, SUM(bandwidth) FROM configs WHERE status='active' AND deleted_at IS NULL AND bandwidth_line_code LIKE '%' GROUP BY bandwidth_line_code, user_code) AS t;")
echo "查询结果: $RESULT 条记录"
if [ "$RESULT" -ge 2 ]; then
    echo "✅ 通过：返回所有用户配置"
else
    echo "❌ 失败：期望至少2条，实际$RESULT条"
    exit 1
fi
echo ""

# 测试6: Grafana API - LIKE '%'
echo "✅ 测试6: Grafana API查询 (LIKE '%')"
echo "--------------------------------------"
RESPONSE=$(curl -s -u admin:admin 'http://localhost:3000/api/ds/query' -H 'Content-Type: application/json' --data-raw '{"queries":[{"datasource":{"type":"mysql","uid":"df7uceqwaqzggc"},"rawSql":"SELECT bandwidth_line_code, total_bandwidth FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL AND bandwidth_line_code LIKE '\''%'\''","format":"table","refId":"A"}]}')

if echo "$RESPONSE" | grep -q "BANDWIDTH_LINE_1996819610904563712"; then
    echo "✅ 通过：Grafana API返回正确数据"
    COUNT=$(echo "$RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(len(d['results']['A']['frames'][0]['data']['values'][0]))" 2>/dev/null || echo "0")
    echo "返回 $COUNT 条带宽线路"
else
    echo "❌ 失败：Grafana API返回异常"
    echo "$RESPONSE" | python3 -m json.tool
    exit 1
fi
echo ""

# 测试7: Grafana API - LIKE 特定线路
echo "✅ 测试7: Grafana API查询 (LIKE '特定线路')"
echo "--------------------------------------"
RESPONSE=$(curl -s -u admin:admin 'http://localhost:3000/api/ds/query' -H 'Content-Type: application/json' --data-raw '{"queries":[{"datasource":{"type":"mysql","uid":"df7uceqwaqzggc"},"rawSql":"SELECT bandwidth_line_code, total_bandwidth FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL AND bandwidth_line_code LIKE '\''BANDWIDTH_LINE_1996819610904563712'\''","format":"table","refId":"A"}]}')

COUNT=$(echo "$RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(len(d['results']['A']['frames'][0]['data']['values'][0]))" 2>/dev/null || echo "0")
if [ "$COUNT" -eq 1 ]; then
    echo "✅ 通过：Grafana API返回特定线路"
    echo "返回线路: BANDWIDTH_LINE_1996819610904563712"
else
    echo "❌ 失败：期望1条，实际${COUNT}条"
    exit 1
fi
echo ""

# 测试8: 验证JSON配置
echo "✅ 测试8: 验证变量和Panel配置"
echo "--------------------------------------"
if grep -q '"allValue": "%"' /home/ubuntu/src/tunnel_monitor/dashboards/business-base.json; then
    echo "✅ 通过：bandwidth_line变量allValue为%"
else
    echo "❌ 失败：allValue未设置为%"
    exit 1
fi

if grep -q 'bandwidth_line_code LIKE' /home/ubuntu/src/tunnel_monitor/dashboards/panels/server/各线路带宽总量.json; then
    echo "✅ 通过：Panel使用LIKE语句"
else
    echo "❌ 失败：Panel未使用LIKE语句"
    exit 1
fi
echo ""

# 最终总结
echo "======================================"
echo "🎉 所有测试通过！"
echo "======================================"
echo ""
echo "解决方案概述："
echo "  ✅ allValue设置为 '%' (SQL通配符)"
echo "  ✅ 所有查询使用 LIKE 语句"
echo "  ✅ 选择All时: bandwidth_line_code LIKE '%' 匹配所有"
echo "  ✅ 选择特定线路: LIKE 'BANDWIDTH_LINE_xxx' 精确匹配"
echo ""
echo "优势："
echo "  • 避免Grafana变量替换时的引号处理问题"
echo "  • SQL语法简洁，无需复杂的OR条件"
echo "  • % 是SQL标准通配符，语义清晰"
echo "  • 完全兼容Grafana的变量机制"
echo ""
echo "后续操作："
echo "  1. 打开 http://localhost:3000"
echo "  2. 进入IPTunnel业务监控面板"
echo "  3. 测试带宽线路下拉框"
echo "     - 选择'All'查看所有数据"
echo "     - 选择特定线路查看该线路数据"
echo ""
