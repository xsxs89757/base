#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$ROOT_DIR/server"
ADMIN_DIR="$ROOT_DIR/admin"
AIR_PID=""
ADMIN_PID=""
ADMIN_PORT="${ADMIN_PORT:-5666}"
FORCE_MODE=0

usage() {
    echo -e "${CYAN}用法: ./dev.sh [--force|-f]${NC}"
    echo ""
    echo "  --force, -f  启动前杀死占用后端/前端开发端口的进程"
    echo "  --help, -h   显示帮助"
    exit "${1:-0}"
}

for arg in "$@"; do
    case "$arg" in
        -f|--force) FORCE_MODE=1 ;;
        -h|--help|help) usage 0 ;;
        *)
            echo -e "${RED}未知参数: $arg${NC}"
            usage 1
            ;;
    esac
done

# 从 config.yaml 读取后端端口（唯一来源）
CONFIG_FILE="$SERVER_DIR/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}config.yaml 不存在，请从 config.yaml.example 复制一份${NC}"
    exit 1
fi
SERVER_PORT=$(grep -E '^\s*port:' "$CONFIG_FILE" | head -1 | awk '{print $2}')
SERVER_PORT=${SERVER_PORT:-8080}

get_port_pids() {
    local port="$1"

    if ! command -v lsof &>/dev/null; then
        return 0
    fi

    lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true
}

collect_kill_pids_for_port() {
    local port="$1"
    local pid ppid parent_cmd

    for pid in $(get_port_pids "$port"); do
        echo "$pid"

        ppid=$(ps -p "$pid" -o ppid= 2>/dev/null | tr -d ' ' || true)
        [ -z "$ppid" ] && continue
        [ "$ppid" = "1" ] && continue
        [ "$ppid" = "$$" ] && continue

        parent_cmd=$(ps -p "$ppid" -o command= 2>/dev/null || true)
        case "$parent_cmd" in
            *"/air"*|*" air"*|*"air "*) echo "$ppid" ;;
        esac
    done | sort -u
}

format_pids() {
    tr '\n' ' ' | sed 's/[[:space:]]*$//'
}

print_processes() {
    local pids="$1"
    local pid cmd port_info

    for pid in $pids; do
        port_info=$(lsof -nP -a -p "$pid" -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR == 2 {print $1 " " $2 " " $9}' || true)
        if [ -n "$port_info" ]; then
            echo -e "        $port_info"
            continue
        fi

        cmd=$(ps -p "$pid" -o command= 2>/dev/null || true)
        if [ -n "$cmd" ]; then
            echo -e "        PID $pid: $cmd"
        else
            echo -e "        PID $pid"
        fi
    done
}

kill_port_listeners() {
    local port="$1"
    local label="$2"
    local pids
    local remaining

    pids=$(collect_kill_pids_for_port "$port" | format_pids)
    [ -z "$pids" ] && return 0

    echo -e "${YELLOW}${label} 端口 ${port} 被占用，强制模式将停止以下进程:${NC}"
    print_processes "$pids"

    for pid in $pids; do
        kill "$pid" 2>/dev/null || true
    done
    sleep 1

    remaining=$(collect_kill_pids_for_port "$port" | format_pids)
    if [ -n "$remaining" ]; then
        echo -e "${YELLOW}      进程仍未退出，执行 kill -9: $remaining${NC}"
        for pid in $remaining; do
            kill -9 "$pid" 2>/dev/null || true
        done
        sleep 1
    fi

    remaining=$(get_port_pids "$port" | format_pids)
    if [ -n "$remaining" ]; then
        echo -e "${RED}无法释放 ${label} 端口 ${port}: $remaining${NC}"
        exit 1
    fi

    echo -e "${GREEN}      ${label} 端口 ${port} 已释放${NC}"
}

ensure_port_available() {
    local port="$1"
    local label="$2"
    local pids

    pids=$(get_port_pids "$port" | format_pids)
    [ -z "$pids" ] && return 0

    if [ "$FORCE_MODE" = "1" ]; then
        kill_port_listeners "$port" "$label"
        return 0
    fi

    echo -e "${RED}${label} 端口 ${port} 已被占用:${NC}"
    print_processes "$pids"
    echo -e "${YELLOW}请先停止占用进程，或使用 ./dev.sh --force / make dev-force 强制释放后启动${NC}"
    exit 1
}

cleanup() {
    echo ""
    echo -e "${YELLOW}正在关闭服务...${NC}"
    [ -n "$AIR_PID" ] && kill "$AIR_PID" 2>/dev/null && echo -e "${GREEN}后端已停止${NC}"
    [ -n "$ADMIN_PID" ] && kill "$ADMIN_PID" 2>/dev/null && echo -e "${GREEN}前端已停止${NC}"
    pkill -P $$ 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - DEV   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""

ensure_port_available "$SERVER_PORT" "后端"
ensure_port_available "$ADMIN_PORT" "前端"

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
echo -e "${YELLOW}[2/2] 启动前端 (http://localhost:${ADMIN_PORT})${NC}"
cd "$ADMIN_DIR"

if [ ! -d node_modules ]; then
    echo -e "${YELLOW}      安装前端依赖...${NC}"
    pnpm install --no-frozen-lockfile
fi

VITE_API_PORT=$SERVER_PORT pnpm dev:antd -- --port "$ADMIN_PORT" &
ADMIN_PID=$!
sleep 3

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   全部服务已启动！${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
echo -e "  前端:    ${CYAN}http://localhost:${ADMIN_PORT}${NC}"
echo -e "  后端:    ${CYAN}http://localhost:${SERVER_PORT}${NC}"
echo -e "  Swagger: ${CYAN}http://localhost:${SERVER_PORT}/swagger/index.html${NC}"
echo ""
echo -e "  默认账号: ${YELLOW}super / 123456${NC}"
echo ""
echo -e "${YELLOW}后端文件修改后自动重新编译 (air 热更新)${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
