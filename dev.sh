#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$ROOT_DIR/server"
WEB_DIR="$ROOT_DIR/web"
AIR_PID=""
WEB_PID=""

cleanup() {
    echo ""
    echo -e "${YELLOW}正在关闭服务...${NC}"
    [ -n "$AIR_PID" ] && kill "$AIR_PID" 2>/dev/null && echo -e "${GREEN}后端已停止${NC}"
    [ -n "$WEB_PID" ] && kill "$WEB_PID" 2>/dev/null && echo -e "${GREEN}前端已停止${NC}"
    pkill -P $$ 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - DEV   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""

# --- 检查 air ---
AIR_BIN="$(go env GOPATH)/bin/air"
if [ ! -f "$AIR_BIN" ]; then
    echo -e "${YELLOW}安装 air (Go 热更新工具)...${NC}"
    go install github.com/air-verse/air@latest
fi

# --- 后端 (air 热更新) ---
echo -e "${YELLOW}[1/2] 启动后端 - air 热更新 (http://localhost:8080)${NC}"
cd "$SERVER_DIR"

if [ ! -f go.sum ]; then
    go mod tidy
fi

# generate swagger docs if swag is installed
SWAG_BIN=$(go env GOPATH)/bin/swag
if [ -f "$SWAG_BIN" ]; then
    echo -e "${YELLOW}      生成 Swagger 文档...${NC}"
    "$SWAG_BIN" init -g main.go -o docs --parseDependency --parseInternal 2>/dev/null || true
fi

"$AIR_BIN" &
AIR_PID=$!
sleep 3

if ! kill -0 "$AIR_PID" 2>/dev/null; then
    echo -e "${RED}后端启动失败！${NC}"
    exit 1
fi
echo -e "${GREEN}      后端启动成功 (air PID: $AIR_PID)${NC}"

# --- 前端 ---
echo -e "${YELLOW}[2/2] 启动前端 (http://localhost:5666)${NC}"
cd "$WEB_DIR"

if [ ! -d node_modules ]; then
    echo -e "${YELLOW}      安装前端依赖...${NC}"
    pnpm install --no-frozen-lockfile
fi

pnpm dev:antd &
WEB_PID=$!
sleep 3

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   全部服务已启动！${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
echo -e "  前端:    ${CYAN}http://localhost:5666${NC}"
echo -e "  后端:    ${CYAN}http://localhost:8080${NC}"
echo -e "  Swagger: ${CYAN}http://localhost:8080/swagger/index.html${NC}"
echo ""
echo -e "  默认账号: ${YELLOW}vben / 123456${NC}"
echo ""
echo -e "${YELLOW}后端文件修改后自动重新编译 (air 热更新)${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
