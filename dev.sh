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
SERVER_PID=""
WEB_PID=""

cleanup() {
    echo ""
    echo -e "${YELLOW}正在关闭服务...${NC}"
    [ -n "$SERVER_PID" ] && kill "$SERVER_PID" 2>/dev/null && echo -e "${GREEN}后端已停止${NC}"
    [ -n "$WEB_PID" ] && kill "$WEB_PID" 2>/dev/null && echo -e "${GREEN}前端已停止${NC}"
    # kill child processes
    pkill -P $$ 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - DEV   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""

# --- 后端 ---
echo -e "${YELLOW}[1/3] 编译后端...${NC}"
cd "$SERVER_DIR"

if [ ! -f go.sum ]; then
    go mod tidy
fi

# generate swagger docs if swag is installed
SWAG_BIN=$(go env GOPATH)/bin/swag
if [ -f "$SWAG_BIN" ]; then
    echo -e "${YELLOW}      生成 Swagger 文档...${NC}"
    "$SWAG_BIN" init -g main.go -o docs --parseDependency --parseInternal 2>/dev/null || true
else
    echo -e "${YELLOW}      swag 未安装，跳过文档生成 (go install github.com/swaggo/swag/cmd/swag@latest)${NC}"
fi

go build -o server .
echo -e "${GREEN}      后端编译完成${NC}"

echo -e "${YELLOW}[2/3] 启动后端 (http://localhost:8080)${NC}"
echo -e "${CYAN}      Swagger: http://localhost:8080/swagger/index.html${NC}"
echo -e "${CYAN}      OpenAPI: http://localhost:8080/swagger/doc.json${NC}"
./server &
SERVER_PID=$!
sleep 2

if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo -e "${RED}后端启动失败！${NC}"
    exit 1
fi
echo -e "${GREEN}      后端启动成功 (PID: $SERVER_PID)${NC}"

# --- 前端 ---
echo -e "${YELLOW}[3/3] 启动前端 (http://localhost:5666)${NC}"
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
echo -e "  OpenAPI: ${CYAN}http://localhost:8080/swagger/doc.json${NC}"
echo ""
echo -e "  默认账号: ${YELLOW}vben / 123456${NC}"
echo ""
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
