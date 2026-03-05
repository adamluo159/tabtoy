#!/bin/bash
# generate.sh
# 使用tabtoy工具生成Go、TypeScript代码和JSON文件
set -e

# 设置项目根目录
PROJECT_ROOT=$(cd $(dirname "$0"); pwd)

# 配置参数
excel_dir="/Users/atat/Documents/block_data/excel"  # Excel文件目录
output_dir="/Users/atat/Documents/block/output"  # 输出目录
json_dir="$output_dir/json"  # JSON输出目录
go_out="$output_dir/table.go"  # Go代码输出路径
ts_out="$output_dir/table.ts"  # TypeScript代码输出路径
package="table"  # 包名

# 设置插件路径
GOBIN=$(go env GOBIN)
if [ -z "$GOBIN" ]; then
    GOBIN="$(go env GOPATH)/bin"
fi

# 将GOBIN添加到PATH，确保能找到工具
export PATH="$GOBIN:$PATH"

# 使用项目根目录的可执行文件
PLUGIN="$PROJECT_ROOT/tabtoy"

# 创建输出目录
echo "📁 Creating output directories..."
mkdir -p "$output_dir"
mkdir -p "$json_dir"

# 清理旧文件
echo "🧹 Cleaning old generated files..."
rm -f "$go_out"
rm -f "$ts_out"
rm -rf "$json_dir"/*
mkdir -p "$json_dir"

echo "🏗️  Checking and building required tools..."

# 检查tabtoy工具
executable=""
if [ -f "$PLUGIN" ]; then
    echo "✅ Found project-local tabtoy"
    executable="$PLUGIN"
elif command -v tabtoy &> /dev/null; then
    echo "✅ Found tabtoy in PATH"
    executable="tabtoy"
else
    echo "⚠️  tabtoy not found, checking if we can build it..."
    # 检查是否可以本地构建
    if [ -f "$PROJECT_ROOT/main.go" ]; then
        echo "Building tabtoy locally..."
        go build -o "$PLUGIN" "$PROJECT_ROOT/main.go"
        if [ -f "$PLUGIN" ]; then
            echo "✅ tabtoy built successfully locally"
            executable="$PLUGIN"
        else
            echo "Error: Failed to build tabtoy locally"
            exit 1
        fi
    else
        echo "Error: tabtoy not found and cannot be built locally"
        echo "建议：1. 确保tabtoy已安装；2. 或确保在tabtoy项目根目录下执行脚本"
        exit 1
    fi
fi

echo "🚀 Generating code..."

# 执行生成命令
if "$executable" -mode=v2 -json_dir="$json_dir" -go_out="$go_out" -ts_out="$ts_out" -package="$package" "$excel_dir"/*.xlsx; then
    echo "🎉 Code generated successfully!"
    echo "📋 Generated files:"
    echo "   - Go code: $go_out"
    echo "   - TypeScript code: $ts_out"
    echo "   - JSON files: $json_dir/"
    echo ""
    echo "📦 Output directory: $output_dir"
    echo "📊 Total JSON files: $(ls -la $json_dir | grep -v "^total" | grep -v "^\.\$" | grep -v "^\.\.\$" | wc -l)"
else
    echo "❌ Failed to generate code"
    exit 1
fi