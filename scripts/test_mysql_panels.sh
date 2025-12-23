#!/bin/bash

# 测试MySQL面板查询的SQL语法
# 验证修复后的条件判断逻辑

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DB_USER="root"
DB_NAME="tunnel"

echo -e "${BLUE}🔍 测试MySQL面板查询${NC}"
echo ""

# 获取第一个带宽线路
bandwidth_line=$(mysql -u $DB_USER $DB_NAME -N -e "SELECT bandwidth_line_code FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL LIMIT 1" 2>&1)

if [ -z "$bandwidth_line" ]; then
    echo -e "${RED}✗ 没有找到活跃的带宽线路${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 使用带宽线路: $bandwidth_line${NC}"
echo ""

# 测试1: 各线路带宽总量 - Query A (All选项)
echo -e "${YELLOW}=== 测试1: 各线路带宽总量 - Query A (All) ===${NC}"
sql_a_all="SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 
  AND deleted_at IS NULL 
  AND ('All' = 'All' OR bandwidth_line_code = 'All')"

echo "SQL: $sql_a_all"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_a_all" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Query A (All) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ Query A (All) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""

# 测试2: 各线路带宽总量 - Query A (特定线路)
echo -e "${YELLOW}=== 测试2: 各线路带宽总量 - Query A (特定线路) ===${NC}"
sql_a_specific="SELECT bandwidth_line_code, total_bandwidth 
FROM bandwidth_lines 
WHERE is_active=1 
  AND deleted_at IS NULL 
  AND ('$bandwidth_line' = 'All' OR bandwidth_line_code = '$bandwidth_line')"

echo "SQL: $sql_a_specific"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_a_specific" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Query A (特定线路) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ Query A (特定线路) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""

# 测试3: 各线路带宽总量 - Query B (All选项)
echo -e "${YELLOW}=== 测试3: 各线路带宽总量 - Query B (All) ===${NC}"
sql_b_all="SELECT bandwidth_line_code, user_code, SUM(bandwidth) 
FROM configs 
WHERE status='active' 
  AND deleted_at IS NULL 
  AND ('All' = 'All' OR bandwidth_line_code = 'All') 
GROUP BY bandwidth_line_code, user_code"

echo "SQL: $sql_b_all"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_b_all" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Query B (All) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ Query B (All) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""

# 测试4: 各线路带宽总量 - Query B (特定线路)
echo -e "${YELLOW}=== 测试4: 各线路带宽总量 - Query B (特定线路) ===${NC}"
sql_b_specific="SELECT bandwidth_line_code, user_code, SUM(bandwidth) 
FROM configs 
WHERE status='active' 
  AND deleted_at IS NULL 
  AND ('$bandwidth_line' = 'All' OR bandwidth_line_code = '$bandwidth_line') 
GROUP BY bandwidth_line_code, user_code"

echo "SQL: $sql_b_specific"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_b_specific" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Query B (特定线路) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ Query B (特定线路) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""

# 测试5: 各线路带宽被购买使用情况 (All选项)
echo -e "${YELLOW}=== 测试5: 各线路带宽被购买使用情况 (All) ===${NC}"
sql_purchase_all="SELECT bandwidth_line_code, user_code, SUM(bandwidth) AS total_purchase 
FROM configs 
WHERE status='active' 
  AND deleted_at IS NULL 
  AND ('All' = 'All' OR bandwidth_line_code = 'All') 
GROUP BY bandwidth_line_code, user_code"

echo "SQL: $sql_purchase_all"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_purchase_all" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 购买使用情况查询 (All) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ 购买使用情况查询 (All) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""

# 测试6: 各线路带宽被购买使用情况 (特定线路)
echo -e "${YELLOW}=== 测试6: 各线路带宽被购买使用情况 (特定线路) ===${NC}"
sql_purchase_specific="SELECT bandwidth_line_code, user_code, SUM(bandwidth) AS total_purchase 
FROM configs 
WHERE status='active' 
  AND deleted_at IS NULL 
  AND ('$bandwidth_line' = 'All' OR bandwidth_line_code = '$bandwidth_line') 
GROUP BY bandwidth_line_code, user_code"

echo "SQL: $sql_purchase_specific"
echo ""

result=$(mysql -u $DB_USER $DB_NAME -e "$sql_purchase_specific" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 购买使用情况查询 (特定线路) 执行成功${NC}"
    echo "$result"
else
    echo -e "${RED}✗ 购买使用情况查询 (特定线路) 执行失败${NC}"
    echo "$result"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ 所有MySQL查询测试通过！${NC}"
echo ""
echo -e "${BLUE}📝 修复总结:${NC}"
echo "  - 使用条件判断代替REGEXP"
echo "  - All选项: ('\$bandwidth_line' = 'All' OR ...) → 第一个条件永远为true"
echo "  - 特定线路: ('line_code' = 'All' OR bandwidth_line_code = 'line_code') → 第二个条件匹配"
echo "  - 避免了 REGEXP 'All' 的SQL语法错误"
echo ""
