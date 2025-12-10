#!/bin/bash

# 运行所有 LangGraphGo 例子的脚本
# 使用方法: ./run_all_examples.sh [timeout_seconds]

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# 配置
TIMEOUT=${1:-60}  # 默认超时时间 60 秒
RESULTS_FILE="example_results.txt"
SUMMARY_FILE="example_summary.txt"

# 清空之前的结果文件
> "$RESULTS_FILE"
> "$SUMMARY_FILE"

# 计数器
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

# 检查是否存在统一的 go.mod
if [ ! -f "go.mod" ]; then
    echo -e "${BLUE}Creating unified go.mod for examples...${NC}"
    cat > go.mod << EOF
module examples

go 1.21

replace github.com/smallnest/langgraphgo => ../

require (
	github.com/smallnest/langgraphgo v0.0.0-00010101000000-000000000000
)
EOF
fi

# 打印标题
echo -e "${BOLD}${BLUE}🚀 LangGraphGo Examples Runner${NC}"
echo -e "${BLUE}=====================================${NC}"
echo -e "Timeout per example: ${TIMEOUT} seconds"
echo -e "Results will be saved to: $RESULTS_FILE"
echo -e "Using unified go.mod in examples directory"
echo

# 获取所有例子目录
EXAMPLE_DIRS=$(find . -maxdepth 1 -type d -not -path '*/\.*' | grep -v "^\.$" | sort)

# 检查是否安装了 Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# 下载统一的依赖
echo -e "${BLUE}Downloading dependencies for all examples...${NC}"
if ! go mod tidy > /dev/null 2>&1; then
    echo -e "${RED}Warning: Failed to download some dependencies${NC}"
fi

# 函数：运行单个例子
run_example() {
    local example_dir=$1
    local example_name=$(basename "$example_dir")

    echo -e "\n${YELLOW}📁 Running: $example_name${NC}"
    echo -e "${YELLOW}$(printf '─%.0s' {1..50})${NC}"

    # 检查是否有 main.go
    if [ ! -f "$example_dir/main.go" ]; then
        echo -e "${RED}❌ No main.go found in $example_name${NC}"
        echo "$example_name: SKIPPED (no main.go)" >> "$RESULTS_FILE"
        ((SKIPPED++))
        return
    fi

    # 检查是否需要 API keys（通过检查代码中的关键词）
    if grep -qiE "(openai.*api.*key|anthropic.*api.*key|tavily.*api.*key|brave.*api.*key|exa.*api.*key)" "$example_dir/main.go" 2>/dev/null; then
        # 检查环境变量是否设置
        if [ -z "$OPENAI_API_KEY" ] && [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$TAVILY_API_KEY" ] && [ -z "$BRAVE_API_KEY" ] && [ -z "$EXA_API_KEY" ]; then
            echo -e "${YELLOW}⚠️  $example_name requires API keys (OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.)${NC}"
            echo "$example_name: SKIPPED (requires API keys)" >> "$RESULTS_FILE"
            ((SKIPPED++))
            return
        fi
    fi

    # 运行例子
    echo -e "🏃 Running..."
    local output_file="/tmp/${example_name}_output.log"
    local error_file="/tmp/${example_name}_error.log"

    # 使用 timeout 命令限制运行时间
    # 兼容 macOS 和 Linux
    if command -v gtimeout &> /dev/null; then
        TIMEOUT_CMD="gtimeout"
    elif command -v timeout &> /dev/null; then
        TIMEOUT_CMD="timeout"
    else
        # 如果没有 timeout 命令，不设置超时限制
        TIMEOUT_CMD=""
    fi

    # 运行命令：指定所有Go文件的路径
    local go_files="$example_dir"/*.go

    if [ -n "$TIMEOUT_CMD" ]; then
        if $TIMEOUT_CMD "$TIMEOUT" go run $go_files > "$output_file" 2> "$error_file"; then
            RUN_STATUS=0
        else
            RUN_STATUS=$?
        fi
    else
        if go run $go_files > "$output_file" 2> "$error_file"; then
            RUN_STATUS=0
        else
            RUN_STATUS=$?
        fi
    fi

    if [ $RUN_STATUS -eq 0 ]; then
        echo -e "${GREEN}✅ $example_name: PASSED${NC}"
        echo "$example_name: PASSED" >> "$RESULTS_FILE"
        ((PASSED++))
    else
        if [ $RUN_STATUS -eq 124 ] && [ -n "$TIMEOUT_CMD" ]; then
            echo -e "${RED}⏱️  $example_name: FAILED (timeout after ${TIMEOUT}s)${NC}"
            echo "$example_name: FAILED (timeout)" >> "$RESULTS_FILE"
        else
            echo -e "${RED}❌ $example_name: FAILED (exit code: $RUN_STATUS)${NC}"
            echo "$example_name: FAILED (exit code: $RUN_STATUS)" >> "$RESULTS_FILE"

            # 显示错误信息的前几行
            if [ -s "$error_file" ]; then
                echo -e "${RED}Error details:${NC}"
                head -10 "$error_file" | sed 's/^/  /'
            fi
        fi
        ((FAILED++))
    fi

    # 清理临时文件
    rm -f "$output_file" "$error_file"
}

# 主循环
for example_dir in $EXAMPLE_DIRS; do
    ((TOTAL++))
    run_example "$example_dir"
done

# 生成总结
echo -e "\n${BOLD}${BLUE}📊 Results Summary${NC}"
echo -e "${BLUE}==================${NC}"

echo "Total examples: $TOTAL" >> "$SUMMARY_FILE"
echo "Passed: $PASSED" >> "$SUMMARY_FILE"
echo "Failed: $FAILED" >> "$SUMMARY_FILE"
echo "Skipped: $SKIPPED" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "Success rate: $(( PASSED * 100 / TOTAL ))%" >> "$SUMMARY_FILE"

echo -e "Total examples: ${BOLD}$TOTAL${NC}"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
echo

if [ $TOTAL -gt 0 ]; then
    success_rate=$(( PASSED * 100 / TOTAL ))
    echo -e "Success rate: ${BOLD}$success_rate%${NC}"
fi

echo
echo -e "${BLUE}Detailed results saved to: $RESULTS_FILE${NC}"
echo -e "${BLUE}Summary saved to: $SUMMARY_FILE${NC}"

# 显示失败的例子
if [ $FAILED -gt 0 ]; then
    echo
    echo -e "${RED}Failed examples:${NC}"
    grep "FAILED" "$RESULTS_FILE" | sed 's/^/  - /'
fi

# 显示跳过的例子
if [ $SKIPPED -gt 0 ]; then
    echo
    echo -e "${YELLOW}Skipped examples:${NC}"
    grep "SKIPPED" "$RESULTS_FILE" | sed 's/^/  - /'
fi

# 根据结果设置退出码
if [ $FAILED -gt 0 ]; then
    exit 1
else
    exit 0
fi