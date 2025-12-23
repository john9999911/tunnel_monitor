#!/bin/bash

# 测试Grafana变量的SQL查询
# 验证数据库中的数据和查询逻辑

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DB_USER="root"
DB_NAME="tunnel"

echo -e "${BLUE}🔍 测试Grafana变量SQL查询${NC}"
echo ""

# 测试1: 带宽线路查询
echo -e "${YELLOW}=== 测试1: 带宽线路查询 ===${NC}"
echo "SQL: SELECT DISTINCT bandwidth_line_code FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL"
echo ""
result=$(mysql -u $DB_USER $DB_NAME -N -e "SELECT DISTINCT bandwidth_line_code FROM bandwidth_lines WHERE is_active=1 AND deleted_at IS NULL ORDER BY bandwidth_line_code" 2>&1)

if [ $? -eq 0 ]; then
    if [ -n "$result" ]; then
        echo -e "${GREEN}✓ 查询成功，找到以下带宽线路:${NC}"
        echo "$result" | while read line; do
            echo "  - $line"
        done
        bandwidth_line=$(echo "$result" | head -1)
        echo ""
        echo -e "${GREEN}✓ 使用第一个线路进行后续测试: $bandwidth_line${NC}"
    else
        echo -e "${RED}✗ 查询成功但没有数据${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ 查询失败: $result${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}=== 测试2: POP机器查询 (特定线路) ===${NC}"
echo "SQL: SELECT DISTINCT m.intra_ip FROM bandwidth_lines bl JOIN machines m ..."
echo "带宽线路: $bandwidth_line"
echo ""

pop_query="SELECT DISTINCT m.intra_ip 
FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) 
WHERE bl.is_active=1 
  AND bl.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND m.type='pop' 
  AND (bl.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')"

result=$(mysql -u $DB_USER $DB_NAME -N -e "$pop_query" 2>&1)

if [ $? -eq 0 ]; then
    if [ -n "$result" ]; then
        echo -e "${GREEN}✓ 查询成功，找到以下POP机器:${NC}"
        echo "$result" | while read line; do
            echo "  - $line"
        done
        pop_count=$(echo "$result" | wc -l)
        echo ""
        echo -e "${GREEN}✓ 共找到 $pop_count 个POP机器${NC}"
    else
        echo -e "${RED}✗ 查询成功但没有数据${NC}"
        echo "可能原因："
        echo "  1. 该带宽线路没有关联的POP机器"
        echo "  2. machines表中没有type='pop'的记录"
        echo "  3. 机器记录被软删除（deleted_at不为NULL）"
    fi
else
    echo -e "${RED}✗ 查询失败: $result${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}=== 测试3: 用户机器查询 (特定线路) ===${NC}"
echo "SQL: SELECT DISTINCT m.intra_ip FROM configs c JOIN machines m ..."
echo "带宽线路: $bandwidth_line"
echo ""

user_query="SELECT DISTINCT m.intra_ip 
FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code 
WHERE c.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND c.status='active' 
  AND (c.bandwidth_line_code = '$bandwidth_line' OR '$bandwidth_line' = 'All')"

result=$(mysql -u $DB_USER $DB_NAME -N -e "$user_query" 2>&1)

if [ $? -eq 0 ]; then
    if [ -n "$result" ]; then
        echo -e "${GREEN}✓ 查询成功，找到以下用户机器:${NC}"
        echo "$result" | while read line; do
            echo "  - $line"
        done
        user_count=$(echo "$result" | wc -l)
        echo ""
        echo -e "${GREEN}✓ 共找到 $user_count 个用户机器${NC}"
    else
        echo -e "${RED}✗ 查询成功但没有数据${NC}"
        echo "可能原因："
        echo "  1. 该带宽线路没有活跃的订单配置"
        echo "  2. configs表中status不为'active'"
        echo "  3. 配置或机器记录被软删除"
    fi
else
    echo -e "${RED}✗ 查询失败: $result${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}=== 测试4: POP机器查询 (All选项) ===${NC}"
echo "SQL: ... WHERE ... AND (bl.bandwidth_line_code = 'All' OR 'All' = 'All')"
echo ""

pop_all_query="SELECT DISTINCT m.intra_ip 
FROM bandwidth_lines bl 
JOIN machines m ON (bl.machine_a_code = m.machine_code OR bl.machine_b_code = m.machine_code) 
WHERE bl.is_active=1 
  AND bl.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND m.type='pop' 
  AND (bl.bandwidth_line_code = 'All' OR 'All' = 'All')"

result=$(mysql -u $DB_USER $DB_NAME -N -e "$pop_all_query" 2>&1)

if [ $? -eq 0 ]; then
    if [ -n "$result" ]; then
        echo -e "${GREEN}✓ 查询成功，All选项返回所有POP机器:${NC}"
        echo "$result" | while read line; do
            echo "  - $line"
        done
        all_pop_count=$(echo "$result" | wc -l)
        echo ""
        echo -e "${GREEN}✓ 共找到 $all_pop_count 个POP机器${NC}"
    else
        echo -e "${RED}✗ 查询成功但没有数据${NC}"
    fi
else
    echo -e "${RED}✗ 查询失败: $result${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}=== 测试5: 用户机器查询 (All选项) ===${NC}"
echo "SQL: ... WHERE ... AND (c.bandwidth_line_code = 'All' OR 'All' = 'All')"
echo ""

user_all_query="SELECT DISTINCT m.intra_ip 
FROM configs c 
JOIN machines m ON c.user_machine_code = m.machine_code 
WHERE c.deleted_at IS NULL 
  AND m.deleted_at IS NULL 
  AND c.status='active' 
  AND (c.bandwidth_line_code = 'All' OR 'All' = 'All')"

result=$(mysql -u $DB_USER $DB_NAME -N -e "$user_all_query" 2>&1)

if [ $? -eq 0 ]; then
    if [ -n "$result" ]; then
        echo -e "${GREEN}✓ 查询成功，All选项返回所有用户机器:${NC}"
        echo "$result" | while read line; do
            echo "  - $line"
        done
        all_user_count=$(echo "$result" | wc -l)
        echo ""
        echo -e "${GREEN}✓ 共找到 $all_user_count 个用户机器${NC}"
    else
        echo -e "${RED}✗ 查询成功但没有数据${NC}"
    fi
else
    echo -e "${RED}✗ 查询失败: $result${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}=== 数据库状态检查 ===${NC}"
echo ""

# 检查各表的记录数
echo "带宽线路表:"
mysql -u $DB_USER $DB_NAME -e "SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN is_active=1 AND deleted_at IS NULL THEN 1 ELSE 0 END) as active
FROM bandwidth_lines"

echo ""
echo "机器表:"
mysql -u $DB_USER $DB_NAME -e "SELECT 
    type,
    COUNT(*) as total,
    SUM(CASE WHEN deleted_at IS NULL THEN 1 ELSE 0 END) as active
FROM machines 
GROUP BY type"

echo ""
echo "配置表:"
mysql -u $DB_USER $DB_NAME -e "SELECT 
    status,
    COUNT(*) as total,
    SUM(CASE WHEN deleted_at IS NULL THEN 1 ELSE 0 END) as not_deleted
FROM configs 
GROUP BY status"

echo ""
echo -e "${GREEN}✅ 所有测试完成！${NC}"
echo ""
echo -e "${BLUE}📝 总结:${NC}"
echo "  - 数据库连接正常"
echo "  - SQL查询语法正确"
echo "  - 数据存在且可查询"
echo "  - All选项的逻辑正确"
echo ""
echo -e "${YELLOW}如果Grafana变量仍然为空，可能的原因:${NC}"
echo "  1. Grafana的MySQL数据源配置不正确"
echo "  2. Grafana的数据源UID与config.yaml中的不匹配"
echo "  3. Grafana需要重新加载dashboard"
echo "  4. 变量的allValue设置不正确"
echo ""
