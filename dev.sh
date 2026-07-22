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
    echo "  默认:        端口被占用时自动改用空闲端口启动 (不影响其他项目)"
    echo "  --force, -f  杀死占用后端/前端开发端口的进程，坚持使用配置端口"
    echo "  --help, -h   显示帮助"
    echo ""
    echo "  仓库根存在 dev.project.sh 时会一并启动下游扩展服务"
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

port_in_use() {
    local port="$1"
    [ -n "$(get_port_pids "$port")" ]
}

# 本次启动已分配出去的端口。多个服务(含 dev.project.sh 扩展服务)依次 resolve_port 时，
# 先解析的服务可能尚未真正 LISTEN，仅靠 lsof 探测会把同一端口分给两个服务
CLAIMED_PORTS=""

port_claimed() {
    case " $CLAIMED_PORTS " in
        *" $1 "*) return 0 ;;
    esac
    return 1
}

find_free_port() {
    local port="$1"
    local limit=$((port + 100))

    while [ "$port" -le "$limit" ]; do
        if ! port_claimed "$port" && ! port_in_use "$port"; then
            echo "$port"
            return 0
        fi
        port=$((port + 1))
    done
    return 1
}

# 解析最终使用的端口，结果写入 RESOLVED_PORT：
#   默认模式  端口被占用 -> 自动挑选空闲端口
#   强制模式  端口被占用 -> 杀死占用进程，坚持使用配置端口
RESOLVED_PORT=""
resolve_port() {
    local port="$1"
    local label="$2"
    local pids free_port

    RESOLVED_PORT="$port"
    pids=$(get_port_pids "$port" | format_pids)

    if [ -z "$pids" ] && ! port_claimed "$port"; then
        CLAIMED_PORTS="$CLAIMED_PORTS $port"
        return 0
    fi

    if [ -n "$pids" ] && [ "$FORCE_MODE" = "1" ]; then
        kill_port_listeners "$port" "$label"
        CLAIMED_PORTS="$CLAIMED_PORTS $port"
        return 0
    fi

    if [ -n "$pids" ]; then
        echo -e "${YELLOW}${label} 端口 ${port} 已被占用:${NC}"
        print_processes "$pids"
    else
        echo -e "${YELLOW}${label} 端口 ${port} 与本次启动的其他服务冲突${NC}"
    fi

    if ! free_port=$(find_free_port $((port + 1))); then
        echo -e "${RED}未找到空闲的${label}端口 (从 ${port} 起已尝试 100 个)${NC}"
        echo -e "${YELLOW}可使用 ./dev.sh --force / make dev-force 强制释放配置端口${NC}"
        exit 1
    fi

    RESOLVED_PORT="$free_port"
    CLAIMED_PORTS="$CLAIMED_PORTS $free_port"
    echo -e "${GREEN}      自动改用空闲端口 ${free_port} (如需固定端口: ./dev.sh --force)${NC}"
}

# 杀掉整棵进程树：pnpm 是多层包装，vite/server 是孙进程，
# 只杀直接子进程会在非交互(kill)场景下留下孤儿监听进程。
# 必须先收集完整棵树再统一 kill——边杀边遍历时上层先退出，
# 下层会被过继给 PID 1，pgrep -P 就找不到了
collect_tree() {
    local p
    echo "$1"
    for p in $(pgrep -P "$1" 2>/dev/null); do
        collect_tree "$p"
    done
}

kill_tree() {
    local pids
    pids=$(collect_tree "$1" | tr '\n' ' ')
    [ -n "$pids" ] && kill $pids 2>/dev/null
    return 0
}

cleanup() {
    echo ""
    echo -e "${YELLOW}正在关闭服务...${NC}"
    [ -n "$AIR_PID" ] && kill_tree "$AIR_PID" && echo -e "${GREEN}后端已停止${NC}"
    [ -n "$ADMIN_PID" ] && kill_tree "$ADMIN_PID" && echo -e "${GREEN}前端已停止${NC}"
    type project_dev_stop &>/dev/null && project_dev_stop
    pkill -P $$ 2>/dev/null
    exit 0
}


trap cleanup SIGINT SIGTERM

# --- 下游挂载点: dev.project.sh (基底不包含此文件、永不创建，下游按需新增) ---
# 在仓库根新增 dev.project.sh 即可挂载额外开发服务，可实现三个函数:
#   project_dev_start  后端/前端启动完成后调用: 用 resolve_port 解析端口
#                      (自动处理占用/强制模式/与已分配端口去重)，后台启动服务并记下 PID
#   project_dev_stop   Ctrl+C 清理时调用: 用 kill_tree <PID> 停掉自己启动的服务
#   project_dev_info   启动汇总里追加打印服务地址行
if [ -f "$ROOT_DIR/dev.project.sh" ]; then
    source "$ROOT_DIR/dev.project.sh"
fi

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - DEV   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""

resolve_port "$SERVER_PORT" "后端"
SERVER_PORT="$RESOLVED_PORT"
resolve_port "$ADMIN_PORT" "前端"
ADMIN_PORT="$RESOLVED_PORT"

# 后端 config.Load 支持 SERVER_PORT 环境变量覆盖 config.yaml，
# air 启动的服务进程会继承该变量，自动换端口才能生效
export SERVER_PORT

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

# 端口通过 VITE_ADMIN_PORT 传递：pnpm 多层转发会把 `-- --port` 吞成位置参数，
# vite 收不到 --port，因此改用环境变量（见 apps/web-antd/vite.config.mts）
VITE_API_PORT=$SERVER_PORT VITE_ADMIN_PORT=$ADMIN_PORT pnpm dev:antd &
ADMIN_PID=$!
sleep 3

# --- 下游扩展服务 (dev.project.sh) ---
if type project_dev_start &>/dev/null; then
    project_dev_start
fi

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   全部服务已启动！${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
echo -e "  前端:    ${CYAN}http://localhost:${ADMIN_PORT}${NC}"
echo -e "  后端:    ${CYAN}http://localhost:${SERVER_PORT}${NC}"
echo -e "  Swagger: ${CYAN}http://localhost:${SERVER_PORT}/swagger/index.html${NC}"
if type project_dev_info &>/dev/null; then
    project_dev_info
fi
echo ""
echo -e "  默认账号: ${YELLOW}super / 123456${NC}"
echo ""
echo -e "${YELLOW}后端文件修改后自动重新编译 (air 热更新)${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
