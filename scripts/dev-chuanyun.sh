#!/bin/bash
# 穿云内网穿透接入，由 dev.sh 在启动早期 source。
#
# 从 dev.sh 里拆出来，纯粹是因为它占了那个脚本三分之一的篇幅（约 260/697 行），
# 把真正的「起前后端」逻辑淹没了。行为与拆分前完全一致：这里只定义变量和函数，
# 调用时机全部留在 dev.sh 原处。
#
# 对下游的承诺不变：dev.project.sh 里照常可以用 chuanyun_up / chuanyun_slug /
# CHUANYUN_LAST_URL —— dev.sh 先 source 本文件，再 source dev.project.sh。
#
# 穿云客户端没装/没开/没登录时全部静默跳过：这是可选便利，不该拦住本地开发。
# 关闭：./dev.sh --no-chuanyun 或 CHUANYUN=0 ./dev.sh
#
# 依赖 dev.sh 已定义: ROOT_DIR、配色变量

# 给后端端口挂一个固定公网地址：微信/支付回调、手机真机联调、演示给别人看。
# 穿云客户端没装/没开/没登录时全部静默跳过——这是可选便利，不该拦住本地开发。
# 关闭：./dev.sh --no-chuanyun 或 CHUANYUN=0 ./dev.sh
CHUANYUN_API_PORT="${CHUANYUN_API_PORT:-7075}"
CHUANYUN_ENABLED=1
[ "${CHUANYUN:-}" = "0" ] && CHUANYUN_ENABLED=0
CHUANYUN_TOML="$ROOT_DIR/chuanyun.toml"   # 可选的项目级隧道配置，见 chuanyun.toml.example
CHUANYUN_LOCAL="$ROOT_DIR/chuanyun.local.toml"   # 个人的：接同事隧道要的口令，不进 git
CHUANYUN_PROJECT=""      # chuanyun.toml 里的 project，用作隧道名前缀
CHUANYUN_CONNECT_PORTS="" # 已建立的 connect 本地端口，退出时断开
CHUANYUN_EXTRA_INFO=""   # chuanyun.toml 带来的额外地址，汇总时打印
CHUANYUN_TUNNELS=""      # 本次登记的所有隧道名，退出时逐个注销（含 dev.project.sh 建的）
CHUANYUN_PUBLIC_URL=""   # 后端公网地址，export 给后端进程拼回调用
CHUANYUN_LAST_URL=""     # 最近一次 chuanyun_up 拿到的地址，供调用方取用

# 解析 chuanyun.toml 的受限子集，输出制表符分隔的记录：
#   project <名字>
#   tunnel  <name> <port>
#   connect <local_port> <from> <auth>
# connect 的 auth 从 chuanyun.local.toml（不进 git）里按 from 匹配补上：
#   [[connects]]
#   from = "zhangsan-api"
#   auth = "user:pass"
# chuanyun.toml 里直接写 auth 也认，但那文件进 git，口令会永远留在历史里。
# 只认 `key = value` 和 `[[section]]`。不用 python/toml 库：Windows Git Bash 未必有 python3，
# 而这个子集用 awk 足够解析
chuanyun_toml_records() {
    [ -f "$CHUANYUN_TOML" ] || return 0
    # 个人口令文件先由 shell 解析成 "from\tauth" 行，通过 -v 传给 awk 拆进数组，
    # 免得 awk 里再嵌一层文件读取
    local local_pairs
    local_pairs=$(chuanyun_local_auths | tr '\n' '\001')
    awk -v local_pairs="$local_pairs" '
        BEGIN {
            n = split(local_pairs, rows, "\001")
            for (r = 1; r <= n; r++) {
                if (rows[r] == "") continue
                t = index(rows[r], "\t")
                if (t > 0) local_auth[substr(rows[r], 1, t - 1)] = substr(rows[r], t + 1)
            }
        }
        function emit() {
            if (sec == "tunnel" && name != "" && port != "")
                printf "tunnel\t%s\t%s\n", name, port
            else if (sec == "connect" && lport != "" && from != "") {
                # 口令优先取 chuanyun.local.toml 里同一个 from 的
                if (from in local_auth) auth = local_auth[from]
                printf "connect\t%s\t%s\t%s\n", lport, from, auth
            }
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

# chuanyun.local.toml → 每行 "<from>\t<auth>"。个人口令文件，只有 [[connects]] 的 from + auth
chuanyun_local_auths() {
    [ -f "$CHUANYUN_LOCAL" ] || return 0
    awk '
        function emit() { if (from != "" && auth != "") printf "%s\t%s\n", from, auth; from = ""; auth = "" }
        { line = ""; inq = 0
          for (i = 1; i <= length($0); i++) { c = substr($0, i, 1); if (c == "\"") inq = !inq; if (c == "#" && !inq) break; line = line c }
          $0 = line }
        /^[ \t]*$/ { next }
        /^[ \t]*\[\[[ \t]*connects[ \t]*\]\]/ { emit(); next }
        /^[ \t]*\[/ { emit(); next }
        /=/ { eq = index($0, "="); key = substr($0, 1, eq - 1); val = substr($0, eq + 1)
              gsub(/^[ \t]+|[ \t]+$/, "", key); gsub(/^[ \t]+|[ \t]+$/, "", val); gsub(/^"|"$/, "", val)
              if (key == "from") from = val; else if (key == "auth") auth = val }
        END { emit() }
    ' "$CHUANYUN_LOCAL"
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
# 退出时把隧道**关掉**，不是删掉。DELETE 会把这条隧道连同用户在穿云客户端里给它设的
# 访问口令一起删掉——下次 ./dev.sh 回来的就是一条没门的隧道，而且没人提醒。
# 关掉之后它留在客户端列表里（开关是关的、口令还在），下次注册时原地打开。
# 真想彻底删，在客户端里删。
chuanyun_forget() {
    [ -z "$1" ] && return 0
    curl -sf -m 3 -X PATCH "http://127.0.0.1:${CHUANYUN_API_PORT}/api/tunnels/$1" \
        -H 'Content-Type: application/json' -d '{"enabled":false}' >/dev/null 2>&1 || true
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

    # 不要先注销再注册。DELETE 会把穿云里这条隧道连同用户在客户端设的访问口令
    # 一起删掉，再 POST 回来就是一条没口令的隧道——门被启动脚本静默拆了。
    # 穿云 0.1.11 起同名重注册是幂等的：端口没变直接成功，端口变了（自动避让）
    # 会关掉重开指向新端口，口令原样保留。
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
