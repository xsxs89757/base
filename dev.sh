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

# 从 config.yaml 读取后端端口（唯一来源）
CONFIG_FILE="$SERVER_DIR/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}config.yaml 不存在，请从 config.yaml.example 复制一份${NC}"
    exit 1
fi
SERVER_PORT=$(grep -E '^\s*port:' "$CONFIG_FILE" | head -1 | awk '{print $2}')
SERVER_PORT=${SERVER_PORT:-8080}

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
echo -e "${YELLOW}[1/2] 启动后端 - air 热更新 (http://localhost:${SERVER_PORT})${NC}"
cd "$SERVER_DIR"

if [ ! -f go.sum ]; then
    go mod tidy
fi

# generate swagger docs
SWAG_BIN=$(go env GOPATH)/bin/swag
if [ -f "$SWAG_BIN" ]; then
    echo -e "${YELLOW}      生成 Swagger 文档...${NC}"
    "$SWAG_BIN" init -g main.go -o docs --parseDependency || true
    echo -e "${GREEN}      Swagger 文档已生成${NC}"
else
    echo -e "${YELLOW}      swag 未安装，跳过文档生成 (go install github.com/swaggo/swag/cmd/swag@latest)${NC}"
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

VITE_API_PORT=$SERVER_PORT pnpm dev:antd &
WEB_PID=$!
sleep 3

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   全部服务已启动！${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
echo -e "  前端:    ${CYAN}http://localhost:5666${NC}"
echo -e "  后端:    ${CYAN}http://localhost:${SERVER_PORT}${NC}"
echo -e "  Swagger: ${CYAN}http://localhost:${SERVER_PORT}/swagger/index.html${NC}"
echo ""
echo -e "  默认账号: ${YELLOW}vben / 123456${NC}"
echo ""
echo -e "${YELLOW}后端文件修改后自动重新编译 (air 热更新)${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
