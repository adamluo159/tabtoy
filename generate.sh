#!/bin/bash

# 生成脚本 - 使用tabtoy工具生成Go、TypeScript代码和JSON文件

# 配置参数
excel_dir="/Users/atat/Documents/block/excel"  # Excel文件目录
output_dir="/Users/atat/Documents/block/output"  # 输出目录
json_dir="$output_dir/json"  # JSON输出目录
go_out="$output_dir/table.go"  # Go代码输出路径
ts_out="$output_dir/table.ts"  # TypeScript代码输出路径
package="table"  # 包名

# 创建输出目录
mkdir -p "$output_dir"
mkdir -p "$json_dir"

# 执行生成命令
echo "开始生成代码..."

# 使用-json_out参数输出单个JSON文件
./tabtoy -mode=v2   -json_dir="$json_dir" -go_out="$go_out" -ts_out="$ts_out" -package="$package" "$excel_dir"/*.xlsx

 