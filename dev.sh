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

# --- Windows 适配（Git Bash / MSYS / Cygwin）---
# Windows 上请在 Git Bash 中运行本脚本；自动切换：
# air 用 .air.windows.toml（无 Unix 环境变量前缀 + .exe 产物）、端口探测用 netstat、杀进程用 taskkill。
IS_WINDOWS=0
case "$(uname -s 2>/dev/null)" in
    MINGW*|MSYS*|CYGWIN*) IS_WINDOWS=1 ;;
esac
EXE=""
[ "$IS_WINDOWS" = "1" ] && EXE=".exe"

# --- 穿云内网穿透（可选）---
# 给后端端口挂一个固定公网地址：微信/支付回调、手机真机联调、演示给别人看。
# 穿云客户端没装/没开/没登录时全部静默跳过——这是可选便利，不该拦住本地开发。
# 关闭：./dev.sh --no-chuanyun 或 CHUANYUN=0 ./dev.sh
CHUANYUN_API_PORT="${CHUANYUN_API_PORT:-7075}"
CHUANYUN_ENABLED=1
[ "${CHUANYUN:-}" = "0" ] && CHUANYUN_ENABLED=0
CHUANYUN_TOML="$ROOT_DIR/chuanyun.toml"   # 可选的项目级隧道配置，见 chuanyun.toml.example
CHUANYUN_PROJECT=""      # chuanyun.toml 里的 project，用作隧道名前缀
CHUANYUN_CONNECT_PORTS="" # 已建立的 connect 本地端口，退出时断开
CHUANYUN_EXTRA_INFO=""   # chuanyun.toml 带来的额外地址，汇总时打印
CHUANYUN_TUNNELS=""      # 本次登记的所有隧道名，退出时逐个注销（含 dev.project.sh 建的）
CHUANYUN_PUBLIC_URL=""   # 后端公网地址，export 给后端进程拼回调用
CHUANYUN_LAST_URL=""     # 最近一次 chuanyun_up 拿到的地址，供调用方取用

usage() {
    echo -e "${CYAN}用法: ./dev.sh [--force|-f]${NC}"
    echo ""
    echo "  默认:        端口被占用时自动改用空闲端口启动 (不影响其他项目)"
    echo "  --force, -f  杀死占用后端/前端开发端口的进程，坚持使用配置端口"
    echo "  --no-chuanyun 本次不接入穿云内网穿透 (等价 CHUANYUN=0)"
    echo "  --help, -h   显示帮助"
    echo ""
    echo "  仓库根存在 dev.project.sh 时会一并启动下游扩展服务"
    echo ""
    echo "  Windows 用户请在 Git Bash 中运行（随 Git for Windows 附带），"
    echo "  脚本会自动改用 .air.windows.toml / netstat / taskkill"
    exit "${1:-0}"
}

for arg in "$@"; do
    case "$arg" in
        -f|--force) FORCE_MODE=1 ;;
        --no-chuanyun) CHUANYUN_ENABLED=0 ;;
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

    if command -v lsof &>/dev/null; then
        lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true
        return 0
    fi

    # Windows(Git Bash) 没有 lsof，用系统 netstat 解析监听 PID
    # 行格式: TCP    0.0.0.0:8080    0.0.0.0:0    LISTENING    1234
    if [ "$IS_WINDOWS" = "1" ] && command -v netstat &>/dev/null; then
        netstat -ano -p tcp 2>/dev/null \
            | awk -v p=":$port" '$1 == "TCP" && $4 == "LISTENING" && substr($2, length($2) - length(p) + 1) == p {print $5}' \
            | sort -u || true
    fi
}

# 跨平台杀单个进程：Windows 用 taskkill（Git Bash 的 kill 杀不动原生 Windows 进程）
kill_pid() {
    local pid="$1" force="$2"

    if [ "$IS_WINDOWS" = "1" ]; then
        taskkill //PID "$pid" //F >/dev/null 2>&1 || true
    elif [ "$force" = "1" ]; then
        kill -9 "$pid" 2>/dev/null || true
    else
        kill "$pid" 2>/dev/null || true
    fi
}

collect_kill_pids_for_port() {
    local port="$1"
    local pid ppid parent_cmd

    for pid in $(get_port_pids "$port"); do
        echo "$pid"

        # Git Bash 的 ps 看不到原生 Windows 进程树，跳过父进程(air)探测
        [ "$IS_WINDOWS" = "1" ] && continue

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
        if command -v lsof &>/dev/null; then
            port_info=$(lsof -nP -a -p "$pid" -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR == 2 {print $1 " " $2 " " $9}' || true)
            if [ -n "$port_info" ]; then
                echo -e "        $port_info"
                continue
            fi
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
        kill_pid "$pid" 0
    done
    sleep 1

    remaining=$(collect_kill_pids_for_port "$port" | format_pids)
    if [ -n "$remaining" ]; then
        echo -e "${YELLOW}      进程仍未退出，强制结束: $remaining${NC}"
        for pid in $remaining; do
            kill_pid "$pid" 1
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

    # Windows: Git Bash 的 pgrep/kill 既看不到也杀不动原生进程树，
    # 用 taskkill //T 让系统连同子进程(server.exe / node)一起结束
    if [ "$IS_WINDOWS" = "1" ]; then
        taskkill //PID "$1" //T //F >/dev/null 2>&1 || true
        return 0
    fi

    pids=$(collect_tree "$1" | tr '\n' ' ')
    [ -n "$pids" ] && kill $pids 2>/dev/null
    return 0
}

# 解析 chuanyun.toml 的受限子集，输出制表符分隔的记录：
#   project <名字>
#   tunnel  <name> <port>
#   connect <local_port> <from> <auth>
# 只认 `key = value` 和 `[[section]]`。不用 python/toml 库：Windows Git Bash 未必有 python3，
# 而这个子集用 awk 足够解析
chuanyun_toml_records() {
    [ -f "$CHUANYUN_TOML" ] || return 0
    awk '
        function emit() {
            if (sec == "tunnel" && name != "" && port != "")
                printf "tunnel\t%s\t%s\n", name, port
            else if (sec == "connect" && lport != "" && from != "")
                printf "connect\t%s\t%s\t%s\n", lport, from, auth
            name = ""; port = ""; lport = ""; from = ""; auth = ""
        }
        {
            # 去掉行内注释，引号里的 # 不算
            line = ""; inq = 0
            for (i = 1; i <= length($0); i++) {
                c = substr($0, i, 1)
                if (c == "\"") inq = !inq
                if (c == "#" && !inq) break
                line = line c
            }
            $0 = line
        }
        /^[ \t]*$/ { next }
        /^[ \t]*\[\[[ \t]*tunnels[ \t]*\]\]/  { emit(); sec = "tunnel";  next }
        /^[ \t]*\[\[[ \t]*connects[ \t]*\]\]/ { emit(); sec = "connect"; next }
        /^[ \t]*\[/ { emit(); sec = ""; next }
        /=/ {
            eq = index($0, "=")
            key = substr($0, 1, eq - 1); val = substr($0, eq + 1)
            gsub(/^[ \t]+|[ \t]+$/, "", key)
            gsub(/^[ \t]+|[ \t]+$/, "", val)
            gsub(/^"|"$/, "", val)
            if (sec == "") { if (key == "project") printf "project\t%s\n", val }
            else if (key == "name")       name  = val
            else if (key == "port")       port  = val
            else if (key == "local_port") lport = val
            else if (key == "from")       from  = val
            else if (key == "auth")       auth  = val
        }
        END { emit() }
    ' "$CHUANYUN_TOML"
}

# 项目标识：优先 chuanyun.toml 的 project，其次 .deploy.env 的 PROJECT_NAME，最后仓库目录名。
# 不 source .deploy.env——那会把 SSH_PASS 等一并带进环境
chuanyun_slug() {
    local slug="$CHUANYUN_PROJECT"
    [ -n "$slug" ] && { echo "$slug"; return 0; }
    if [ -f "$ROOT_DIR/.deploy.env" ]; then
        slug=$(grep -E '^[[:space:]]*PROJECT_NAME=' "$ROOT_DIR/.deploy.env" 2>/dev/null \
            | head -1 | cut -d= -f2- | tr -d '"'"'"'[:space:]' || true)
    fi
    [ -z "$slug" ] && slug=$(basename "$ROOT_DIR")
    echo "$slug"
}

# 注销指定隧道，失败无所谓
chuanyun_forget() {
    [ -z "$1" ] && return 0
    curl -sf -m 3 -X DELETE "http://127.0.0.1:${CHUANYUN_API_PORT}/api/tunnels/$1" >/dev/null 2>&1 || true
    return 0
}

# 登记隧道名，退出时统一注销（vite 插件自己建的隧道也走这里登记）
chuanyun_track() {
    case " $CHUANYUN_TUNNELS " in
        *" $1 "*) return 0 ;;
    esac
    CHUANYUN_TUNNELS="$CHUANYUN_TUNNELS $1"
    return 0
}

# 给 <port> 建一条名为 <name> 的隧道，地址写入 CHUANYUN_LAST_URL（没接上则为空）。
# dev.project.sh 里的下游服务可直接复用：
#     chuanyun_up "$(chuanyun_slug)-web" "$WEB_PORT"; WEB_URL="$CHUANYUN_LAST_URL"
# 任何一步失败都只是没有公网地址，不影响本地开发
chuanyun_up() {
    local name="$1" port="$2" url resp base="http://127.0.0.1:${CHUANYUN_API_PORT}"
    CHUANYUN_LAST_URL=""

    # --no-chuanyun / CHUANYUN=0 时一律不建隧道。守卫放在这里而不是调用处，
    # dev.project.sh 里的下游服务照抄示例也自动遵守这个开关
    [ "$CHUANYUN_ENABLED" = "1" ] || return 0

    # 穿云没在跑就安静退出
    curl -sf -m 2 "$base/api/status" >/dev/null 2>&1 || return 0

    # 先注销同名隧道再注册：dev.sh 会自动换端口，上次残留的隧道可能指向旧端口；
    # 且穿云对已存在的名字会直接报"名称已被占用"而不是改端口
    chuanyun_forget "$name"
    # 地址直接取注册响应里的 url。不能按端口 resolve——同一端口可能挂着多条隧道
    # （比如手工建过一条，或 chuanyun.toml 里声明了同端口的另一个名字），那样会取错别人的地址
    resp=$(curl -sf -m 5 -X POST "$base/api/tunnels" -H 'Content-Type: application/json' \
        -d "[{\"port\":${port},\"name\":\"${name}\"}]" 2>/dev/null || true)
    url=$(printf '%s' "$resp" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p' | head -1)

    # 注册没拿到 url（响应异常等）时退回按端口问一次
    [ -z "$url" ] && url=$(curl -sf -m 3 "$base/api/resolve?port=${port}&plain=1" 2>/dev/null || true)
    # 没有隧道时 resolve 会回落成 127.0.0.1，那不算接入成功
    case "$url" in
        http*://127.0.0.1*|http*://localhost*|"") return 0 ;;
        http*://*) chuanyun_track "$name"; CHUANYUN_LAST_URL="$url" ;;
    esac
    return 0
}

# 建立 chuanyun.toml 声明的 [[connects]]：把同事的服务接到本机端口上。
# 早于端口解析执行——connect 会真占住本地端口，后面 resolve_port 自然会避开它
chuanyun_connects_up() {
    local kind lport from auth res body
    [ "$CHUANYUN_ENABLED" = "1" ] || return 0
    [ -f "$CHUANYUN_TOML" ] || return 0
    curl -sf -m 2 "http://127.0.0.1:${CHUANYUN_API_PORT}/api/status" >/dev/null 2>&1 || return 0

    while IFS="$(printf '\t')" read -r kind lport from auth; do
        [ "$kind" = "connect" ] || continue
        body="{\"local_port\":${lport},\"from\":\"${from}\""
        [ -n "$auth" ] && body="${body},\"auth\":\"${auth}\""
        body="${body}}"
        res=$(curl -sf -m 8 -X POST "http://127.0.0.1:${CHUANYUN_API_PORT}/api/connects" \
              -H 'Content-Type: application/json' -d "$body" 2>/dev/null || true)
        case "$res" in
            *'"ok":true'*)
                CHUANYUN_CONNECT_PORTS="$CHUANYUN_CONNECT_PORTS $lport"
                echo -e "${GREEN}[穿云] 已接入 ${from} -> 本地 ${lport}${NC}"
                ;;
            *)
                # 接不上不该拦住开发：本地端口被占、同事没开隧道、口令不对都归到这里
                echo -e "${YELLOW}[穿云] 接入 ${from} 失败，跳过: ${res:-无响应}${NC}"
                ;;
        esac
    done < <(chuanyun_toml_records)
    return 0
}

# 建立 chuanyun.toml 声明的 [[tunnels]]。这里的 port 按原样使用（不经 resolve_port）：
# 声明的语义是"把已经跑在这个端口上的东西暴露出去"，换成别的空闲端口就指向空气了。
# dev.sh 自己启动的服务端口是动态的，那些用 chuanyun_up 现取现用，别写进 toml
chuanyun_tunnels_up() {
    local kind name port
    [ "$CHUANYUN_ENABLED" = "1" ] || return 0
    [ -f "$CHUANYUN_TOML" ] || return 0

    while IFS="$(printf '\t')" read -r kind name port; do
        [ "$kind" = "tunnel" ] || continue
        chuanyun_up "$(chuanyun_slug)-${name}" "$port"
        if [ -n "$CHUANYUN_LAST_URL" ]; then
            CHUANYUN_EXTRA_INFO="${CHUANYUN_EXTRA_INFO}  ${name}(:${port}): ${CHUANYUN_LAST_URL}\n"
        fi
    done < <(chuanyun_toml_records)
    return 0
}

chuanyun_down() {
    local t p
    # 含前端隧道：那条虽是 vite 插件建的，但插件的退出钩子在被 kill 树杀时来不及跑，
    # 名字又是 dev.sh 注入的，这里一并收尾
    for t in $CHUANYUN_TUNNELS; do
        chuanyun_forget "$t"
    done
    for p in $CHUANYUN_CONNECT_PORTS; do
        curl -sf -m 3 -X DELETE "http://127.0.0.1:${CHUANYUN_API_PORT}/api/connects/$p" >/dev/null 2>&1 || true
    done
    return 0
}

cleanup() {
    echo ""
    echo -e "${YELLOW}正在关闭服务...${NC}"
    chuanyun_down
    [ -n "$AIR_PID" ] && kill_tree "$AIR_PID" && echo -e "${GREEN}后端已停止${NC}"
    [ -n "$ADMIN_PID" ] && kill_tree "$ADMIN_PID" && echo -e "${GREEN}前端已停止${NC}"
    type project_dev_stop &>/dev/null && project_dev_stop
    command -v pkill &>/dev/null && pkill -P $$ 2>/dev/null

    # Windows 下杀 bash 作业不一定连带结束原生子进程，按本次已分配的端口兜底清一遍
    # (CLAIMED_PORTS 覆盖后端/前端及 dev.project.sh 扩展服务)
    if [ "$IS_WINDOWS" = "1" ]; then
        local p pid
        for p in $CLAIMED_PORTS; do
            for pid in $(get_port_pids "$p"); do
                kill_pid "$pid" 1
            done
        done
    fi
    exit 0
}


trap cleanup SIGINT SIGTERM

# --- 下游挂载点: dev.project.sh (基底不包含此文件、永不创建，下游按需新增) ---
# 在仓库根新增 dev.project.sh 即可挂载额外开发服务，可实现三个函数:
#   project_dev_start  后端/前端启动完成后调用: 用 resolve_port 解析端口
#                      (自动处理占用/强制模式/与已分配端口去重)，后台启动服务并记下 PID
#   project_dev_stop   Ctrl+C 清理时调用: 用 kill_tree <PID> 停掉自己启动的服务
#   project_dev_info   启动汇总里追加打印服务地址行
#
# ⚠️ 启动时机：**要给后端进程注入环境变量的服务，别放 project_dev_start**——
# 本文件在此处 source，而 project_dev_start 是在后端 air 起来之后才调用的，
# 那时再 export 后端已经看不到了（export 只影响之后 fork 的子进程）。这类服务
# 直接写在 dev.project.sh 的顶层（源载即执行）：那时 resolve_port / kill_tree
# 等助手已就绪，唯独后端端口尚未解析（在本段之后），需要它的服务才留给
# project_dev_start。
if [ -f "$ROOT_DIR/dev.project.sh" ]; then
    source "$ROOT_DIR/dev.project.sh"
fi

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - DEV   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""

# chuanyun.toml：先读出 project（隧道名前缀要用），再建 connects。
# connects 会占住本地端口，所以必须早于 resolve_port，后面才会自动避开
if [ "$CHUANYUN_ENABLED" = "1" ] && [ -f "$CHUANYUN_TOML" ]; then
    CHUANYUN_PROJECT=$(chuanyun_toml_records | awk -F"$(printf '\t')" '$1 == "project" { print $2; exit }')
    chuanyun_connects_up
fi

resolve_port "$SERVER_PORT" "后端"
SERVER_PORT="$RESOLVED_PORT"
resolve_port "$ADMIN_PORT" "前端"
ADMIN_PORT="$RESOLVED_PORT"

# 后端 config.Load 支持 SERVER_PORT 环境变量覆盖 config.yaml，
# air 启动的服务进程会继承该变量，自动换端口才能生效
export SERVER_PORT

# 穿云接入要在启动后端之前完成：CHUANYUN_PUBLIC_URL 得先 export 出去，
# air 之后 fork 的后端进程才拿得到（业务代码可用它拼微信/支付回调地址）
if [ "$CHUANYUN_ENABLED" = "1" ]; then
    CHUANYUN_WEB_TUNNEL="$(chuanyun_slug)-admin"
    # 先清掉上次残留的前端隧道：插件是直接 POST 注册的，撞上同名会报"名称已被占用"，
    # 结果公网地址仍指向上一次那个已经关掉的端口
    chuanyun_forget "$CHUANYUN_WEB_TUNNEL"
    chuanyun_track "$CHUANYUN_WEB_TUNNEL"           # 插件建的隧道也由 dev.sh 负责注销
    chuanyun_up "$(chuanyun_slug)-api" "$SERVER_PORT"
    CHUANYUN_PUBLIC_URL="$CHUANYUN_LAST_URL"
    chuanyun_tunnels_up
    export CHUANYUN_NAME="$CHUANYUN_WEB_TUNNEL"     # 前端隧道名，vite 插件读它
else
    export CHUANYUN=0                                # 让 vite 插件一并跳过
fi
export CHUANYUN_PUBLIC_URL

# --- 检查 air ---
# go env GOPATH 在 Windows 返回反斜杠路径，统一成正斜杠供 bash 使用
GOPATH_DIR="$(go env GOPATH | tr '\\' '/')"
AIR_BIN="$GOPATH_DIR/bin/air$EXE"
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

# generate swagger docs（docs 已随仓库提交，此步只为保持与代码同步；失败不影响启动）
# swag 每次要刷 600+ 行流水账（每个生成的类型一行 "Generating x"，每个它读不懂的
# 第三方类型一行 "TypeSpecDef is nil"），把「route ... declared multiple times」
# 这类真该处理的警告冲得看不见。这里只滤掉流水账——不用 swag 自己的 -q，那个把
# 警告和报错也一并吞了。
SWAG_NOISE='Generating |TypeSpecDef is nil|Generate swagger docs|Generate general API Info|create (docs\.go|swagger\.json|swagger\.yaml) at '
SWAG_BIN="$GOPATH_DIR/bin/swag$EXE"
# docs 缺失（如被误删）且 swag 未装时必须现装现生成，否则 main.go 的 base/docs import 编译不过
if [ ! -f "$SWAG_BIN" ] && [ ! -f docs/docs.go ]; then
    echo -e "${YELLOW}      docs/ 缺失且 swag 未安装，安装 swag (Go Swagger 生成工具)...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
fi
if [ -f "$SWAG_BIN" ]; then
    echo -e "${YELLOW}      生成 Swagger 文档...${NC}"
    SWAG_LOG=$(mktemp)
    if "$SWAG_BIN" init -g main.go -o docs --parseDependency >"$SWAG_LOG" 2>&1; then
        grep -vE "$SWAG_NOISE" "$SWAG_LOG" | sed 's/^/      /' || true
        echo -e "${GREEN}      Swagger 文档已生成${NC}"
    else
        # 失败时不过滤：报错往往就藏在被滤掉的那类行的上下文里
        echo -e "${YELLOW}      Swagger 生成失败（不影响启动，继续用仓库内已有 docs）:${NC}"
        tail -20 "$SWAG_LOG" | sed 's/^/      /'
        # swag 中途失败可能把 docs/ 写坏（docs.go 缺失会让 go build 编译不过、air 起不来），从 git 恢复
        if [ ! -f docs/docs.go ]; then
            echo -e "${YELLOW}      docs/docs.go 缺失，从 git 恢复 docs/ ...${NC}"
            git checkout -- docs 2>/dev/null || true
        fi
    fi
    rm -f "$SWAG_LOG"
else
    echo -e "${YELLOW}      swag 未安装，跳过文档生成 (go install github.com/swaggo/swag/cmd/swag@latest)${NC}"
fi

# Windows 用专用配置：.air.toml 的 build cmd 带 Unix 内联环境变量前缀(CGO_LDFLAGS=-w)，
# PowerShell/cmd 不支持该语法，且产物需要 .exe 后缀
if [ "$IS_WINDOWS" = "1" ]; then
    "$AIR_BIN" -c .air.windows.toml &
else
    "$AIR_BIN" &
fi
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
if [ -n "$CHUANYUN_PUBLIC_URL" ]; then
echo -e "  公网:    ${CYAN}${CHUANYUN_PUBLIC_URL}${NC} (穿云 -> 后端 ${SERVER_PORT})"
fi
if [ -n "$CHUANYUN_EXTRA_INFO" ]; then
printf "%b" "${CYAN}${CHUANYUN_EXTRA_INFO}${NC}"
fi
if type project_dev_info &>/dev/null; then
    project_dev_info
fi
echo ""
echo -e "  默认账号: ${YELLOW}super / 123456${NC}"
echo ""
echo -e "${YELLOW}后端文件修改后自动重新编译 (air 热更新)${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"

wait
