#!/bin/bash

# OpenAPI Registry MCP Service 构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Go 版本
go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | cut -c3-)
required_version=1.26

if [ $(echo "$go_version >= $required_version" | bc -l) -eq 1 ]; then
    print_status "Go 版本 $go_version 满足要求 (>= $required_version)"
else
    print_error "Go 版本 $go_version 过低，需要 >= $required_version"
    exit 1
fi

# 函数定义
build() {
    print_status "正在构建项目..."
    go build -o api-registry-mcp main.go
    print_status "构建完成！可执行文件: ./api-registry-mcp"
}

run() {
    print_status "正在启动服务..."
    print_status "服务将在 http://localhost:8080 启动"
    print_status "按 Ctrl+C 停止服务"
    echo ""
    go run main.go
}

test() {
    print_status "运行测试..."
    go test ./...
}

clean() {
    print_status "清理构建文件..."
    rm -f api-registry-mcp
    print_status "清理完成"
}

help() {
    echo "用法: $0 {build|run|test|clean|all}"
    echo ""
    echo "命令:"
    echo "  build    构建可执行文件"
    echo "  run      运行服务（开发模式）"
    echo "  test     运行测试"
    echo "  clean    清理构建文件"
    echo "  all      构建并运行服务"
    echo "  help     显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 build    # 构建项目"
    echo "  $0 run      # 运行服务"
    echo "  $0 all      # 构建并运行"
}

# 主逻辑
case "$1" in
    build)
        build
        ;;
    run)
        run
        ;;
    test)
        test
        ;;
    clean)
        clean
        ;;
    all)
        build
        echo ""
        run
        ;;
    help|*)
        help
        ;;
esac
