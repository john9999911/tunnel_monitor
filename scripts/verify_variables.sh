#!/bin/bash

# Grafana变量验证脚本
# 用于验证 $pop_machines 和 $user_machines 变量配置是否正确

set -e

echo "🔍 验证Grafana变量配置..."
echo ""

# 配置
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
GRAFANA_USER="${GRAFANA_USER:-admin}"
GRAFANA_PASS="${GRAFANA_PASS:-admin}"
DASHBOARD_UID="iptunnel-business"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "📋 检查配置文件..."

# 检查 business-base.json
if [ -f "./dashboards/business-base.json" ]; then
    echo -e "${GREEN}✓${NC} business-base.json 存在"
    
    # 检查 allValue 字段
    if grep -q '"allValue"' "./dashboards/business-base.json"; then
        echo -e "${GREEN}✓${NC} 发现 allValue 字段"
    else
        echo -e "${RED}✗${NC} 缺少 allValue 字段"
        exit 1
    fi
    
    # 检查 pop_machines 变量
    if grep -q '"name": "pop_machines"' "./dashboards/business-base.json"; then
        echo -e "${GREEN}✓${NC} pop_machines 变量已定义"
    else
        echo -e "${RED}✗${NC} 缺少 pop_machines 变量"
        exit 1
    fi
    
    # 检查 user_machines 变量
    if grep -q '"name": "user_machines"' "./dashboards/business-base.json"; then
        echo -e "${GREEN}✓${NC} user_machines 变量已定义"
    else
        echo -e "${RED}✗${NC} 缺少 user_machines 变量"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} business-base.json 不存在"
    exit 1
fi

echo ""
echo "📋 检查模板文件..."

# 检查 business-template.json
if [ -f "./dashboards/business-template.json" ]; then
    echo -e "${GREEN}✓${NC} business-template.json 存在"
    
    # 检查变量数量
    var_count=$(grep -c '"name": "bandwidth_line"\|"name": "pop_machines"\|"name": "user_machines"' "./dashboards/business-template.json" || true)
    if [ "$var_count" -eq 3 ]; then
        echo -e "${GREEN}✓${NC} 3个变量都已定义"
    else
        echo -e "${YELLOW}⚠${NC} 发现 $var_count 个变量（期望3个）"
    fi
else
    echo -e "${RED}✗${NC} business-template.json 不存在"
    exit 1
fi

echo ""
echo "🔧 验证变量配置详情..."

# 提取并验证 pop_machines 配置
echo ""
echo "--- pop_machines 变量 ---"
if python3 -c "
import json
import sys

try:
    with open('./dashboards/business-base.json', 'r') as f:
        data = json.load(f)
    
    variables = data.get('templating', {}).get('list', [])
    pop_machines = next((v for v in variables if v.get('name') == 'pop_machines'), None)
    
    if not pop_machines:
        print('❌ 未找到 pop_machines 变量')
        sys.exit(1)
    
    print(f\"✓ includeAll: {pop_machines.get('includeAll', False)}\")
    print(f\"✓ multi: {pop_machines.get('multi', False)}\")
    print(f\"✓ allValue: {pop_machines.get('allValue', 'NOT SET')}\")
    print(f\"✓ refresh: {pop_machines.get('refresh', 0)}\")
    print(f\"✓ hide: {pop_machines.get('hide', 0)}\")
    
    # 检查必需字段
    if not pop_machines.get('includeAll'):
        print('⚠ includeAll 应该设置为 true')
        sys.exit(1)
    
    if pop_machines.get('allValue') != '.*':
        print('⚠ allValue 应该设置为 \".*\"')
        sys.exit(1)
    
    print('✅ pop_machines 配置正确')
    
except Exception as e:
    print(f'❌ 错误: {e}')
    sys.exit(1)
" 2>/dev/null; then
    echo -e "${GREEN}pop_machines 验证通过${NC}"
else
    echo -e "${RED}pop_machines 验证失败${NC}"
    exit 1
fi

# 提取并验证 user_machines 配置
echo ""
echo "--- user_machines 变量 ---"
if python3 -c "
import json
import sys

try:
    with open('./dashboards/business-base.json', 'r') as f:
        data = json.load(f)
    
    variables = data.get('templating', {}).get('list', [])
    user_machines = next((v for v in variables if v.get('name') == 'user_machines'), None)
    
    if not user_machines:
        print('❌ 未找到 user_machines 变量')
        sys.exit(1)
    
    print(f\"✓ includeAll: {user_machines.get('includeAll', False)}\")
    print(f\"✓ multi: {user_machines.get('multi', False)}\")
    print(f\"✓ allValue: {user_machines.get('allValue', 'NOT SET')}\")
    print(f\"✓ refresh: {user_machines.get('refresh', 0)}\")
    print(f\"✓ hide: {user_machines.get('hide', 0)}\")
    
    # 检查必需字段
    if not user_machines.get('includeAll'):
        print('⚠ includeAll 应该设置为 true')
        sys.exit(1)
    
    if user_machines.get('allValue') != '.*':
        print('⚠ allValue 应该设置为 \".*\"')
        sys.exit(1)
    
    print('✅ user_machines 配置正确')
    
except Exception as e:
    print(f'❌ 错误: {e}')
    sys.exit(1)
" 2>/dev/null; then
    echo -e "${GREEN}user_machines 验证通过${NC}"
else
    echo -e "${RED}user_machines 验证失败${NC}"
    exit 1
fi

echo ""
echo "✅ 所有配置验证通过！"
echo ""
echo "📝 下一步："
echo "   1. 运行: go run main.go dashboard create"
echo "   2. 在Grafana中访问: ${GRAFANA_URL}/d/${DASHBOARD_UID}"
echo "   3. 检查变量下拉框是否显示正确"
echo "   4. 测试切换不同带宽线路，验证面板数据更新"
echo ""
