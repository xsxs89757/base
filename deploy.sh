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
# BUILD_DIR 在解析完项目名后按项目设置 (.build/<项目名>)

# -------------------------------------------------------
# 使用说明
# -------------------------------------------------------
usage() {
    echo -e "${CYAN}用法: ./deploy.sh [server|admin|all] [项目名]${NC}"
    echo ""
    echo "  server  - 仅部署后端"
    echo "  admin   - 仅部署后台前端"
    echo "  all     - 全量部署 (默认)"
    echo ""
    echo "  项目名  - 可选。指定后读取 .deploy.<项目名>.env，"
    echo "            用于同一台服务器发布多个项目；不指定读取 .deploy.env"
    echo ""
    echo "  -p, --project <项目名>  同上，显式指定项目"
    echo "  -l, --list              列出已有的部署配置"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh                    # 默认配置全量部署"
    echo "  ./deploy.sh server             # 默认配置仅部署后端"
    echo "  ./deploy.sh all shop           # 使用 .deploy.shop.env 全量部署"
    echo "  ./deploy.sh admin -p blog      # 使用 .deploy.blog.env 部署前端"
    exit 0
}

list_projects() {
    echo -e "${CYAN}可用部署配置:${NC}"
    local found=0 f name
    if [ -f "$ROOT_DIR/.deploy.env" ]; then
        echo "  (默认)        .deploy.env"
        found=1
    fi
    for f in "$ROOT_DIR"/.deploy.*.env; do
        [ -f "$f" ] || continue
        name=$(basename "$f")
        name=${name#.deploy.}
        name=${name%.env}
        printf '  %-12s  %s\n' "$name" "$(basename "$f")"
        found=1
    done
    if [ "$found" = "0" ]; then
        echo -e "  ${YELLOW}(暂无，复制 .deploy.env.example 创建)${NC}"
    fi
}

DEPLOY_MODE=""
PROJECT=""
while [ $# -gt 0 ]; do
    case "$1" in
        server|admin|all) DEPLOY_MODE="$1" ;;
        web) DEPLOY_MODE="admin" ;;
        -p|--project)
            shift
            [ -z "${1:-}" ] && { echo -e "${RED}-p/--project 需要指定项目名${NC}"; exit 1; }
            PROJECT="$1"
            ;;
        -l|--list|list) list_projects; exit 0 ;;
        -h|--help|help) usage ;;
        -*) echo -e "${RED}未知参数: $1${NC}"; usage ;;
        *)
            if [ -z "$PROJECT" ]; then
                PROJECT="$1"
            else
                echo -e "${RED}多余参数: $1${NC}"; usage
            fi
            ;;
    esac
    shift
done
DEPLOY_MODE="${DEPLOY_MODE:-all}"

# -------------------------------------------------------
# 加载配置（支持多项目: .deploy.<项目名>.env）
# -------------------------------------------------------
if [ -n "$PROJECT" ]; then
    ENV_FILE="$ROOT_DIR/.deploy.${PROJECT}.env"
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${RED}缺少项目部署配置文件 .deploy.${PROJECT}.env${NC}"
        echo -e "${YELLOW}请复制 .deploy.env.example 为 .deploy.${PROJECT}.env 并填写实际配置${NC}"
        echo ""
        echo "  cp .deploy.env.example .deploy.${PROJECT}.env"
        echo ""
        list_projects
        exit 1
    fi
else
    ENV_FILE="$ROOT_DIR/.deploy.env"
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${RED}缺少部署配置文件 .deploy.env${NC}"
        echo -e "${YELLOW}请复制 .deploy.env.example 为 .deploy.env 并填写实际配置${NC}"
        echo ""
        echo "  cp .deploy.env.example .deploy.env"
        exit 1
    fi
fi

source "$ENV_FILE"

# 项目名解析: .deploy.env 的 PROJECT_NAME > 命令行项目名 > 由 SERVICE_NAME 推导。
# 服务名、远程临时文件、本地构建目录都按项目隔离，同一台服务器可发布多个项目。
# 三者都缺时拒绝部署：base 的每份项目拷贝如果都落到同一个默认服务名，
# 后部署的项目会顶掉先部署项目的 systemd 服务
PROJECT_NAME="${PROJECT_NAME:-$PROJECT}"
if [ -z "$PROJECT_NAME" ] && [ -n "${SERVICE_NAME:-}" ]; then
    PROJECT_NAME="${SERVICE_NAME%-server}"
fi
if [ -z "$PROJECT_NAME" ]; then
    echo -e "${RED}缺少项目标识：多个项目发布到同一台服务器时必须能区分项目${NC}"
    echo -e "${YELLOW}请在 $(basename "$ENV_FILE") 中添加一行（改成你的项目名）:${NC}"
    echo ""
    echo "  PROJECT_NAME=shop"
    echo ""
    echo -e "${YELLOW}或设置 SERVICE_NAME；也可用 ./deploy.sh all <项目名> 走 .deploy.<项目名>.env${NC}"
    exit 1
fi

REMOTE_ADMIN_DIR="${REMOTE_ADMIN_DIR:-${REMOTE_WEB_DIR:-}}"
ADMIN_BUILD_CMD="${ADMIN_BUILD_CMD:-${WEB_BUILD_CMD:-pnpm build:antd}}"

for var in SSH_HOST SSH_USER SSH_PASS REMOTE_SERVER_DIR REMOTE_ADMIN_DIR; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}配置项 $var 未设置，请检查 .deploy.env${NC}"
        exit 1
    fi
done

SSH_PORT="${SSH_PORT:-22}"
AUTO_RESTART="${AUTO_RESTART:-no}"
SERVICE_NAME="${SERVICE_NAME:-${PROJECT_NAME}-server}"
SERVER_BIN_NAME="${SERVER_BIN_NAME:-server}"

# 本地构建目录和远程临时包按项目区分，避免多项目互相覆盖
BUILD_DIR="$ROOT_DIR/.build/${PROJECT_NAME}"
REMOTE_TMP_TAR="/tmp/${PROJECT_NAME}-admin.tar.gz"

# 检测 sshpass
if ! command -v sshpass &>/dev/null; then
    echo -e "${YELLOW}安装 sshpass...${NC}"
    if [[ "$(uname)" == "Darwin" ]]; then
        brew install hudochenkov/sshpass/sshpass
    else
        sudo apt-get install -y sshpass 2>/dev/null || sudo yum install -y sshpass 2>/dev/null
    fi
fi

# 端口参数不能放进公共 SSH_OPTS：ssh 用 -p，而 scp 的 -p 是"保留时间戳"(不带参数)，
# 端口号会被 scp 当成待上传的本地文件 (scp: stat local "22": No such file)
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10"
ssh_run() { sshpass -p "$SSH_PASS" ssh $SSH_OPTS -p "${SSH_PORT}" "${SSH_USER}@${SSH_HOST}" "$@"; }
scp_to() { sshpass -p "$SSH_PASS" scp $SSH_OPTS -P "${SSH_PORT}" "$1" "${SSH_USER}@${SSH_HOST}:$2"; }

# -------------------------------------------------------
# 远程目录归属校验：目录内 .deploy-project 记录属主项目，
# 防止两个项目的配置误指向同一个远程目录而互相覆盖
# -------------------------------------------------------
check_remote_owner() {
    local dir="$1" label="$2" owner
    owner=$(ssh_run "cat ${dir}/.deploy-project 2>/dev/null" | tr -d '[:space:]' || true)
    if [ -n "$owner" ] && [ "$owner" != "$PROJECT_NAME" ]; then
        echo -e "${RED}[${label}] 远程目录 ${dir} 属于项目 '${owner}'，当前项目为 '${PROJECT_NAME}'${NC}"
        echo -e "${YELLOW}多项目共用服务器时请为每个项目配置独立的远程目录；${NC}"
        echo -e "${YELLOW}确认目录确实归本项目使用，可删除该目录下的 .deploy-project 后重试${NC}"
        exit 1
    fi
}

mark_remote_owner() {
    ssh_run "echo '${PROJECT_NAME}' > $1/.deploy-project"
}

# -------------------------------------------------------
# 自动检测远程系统和架构
# -------------------------------------------------------
echo -e "${YELLOW}检测远程服务器系统信息...${NC}"
REMOTE_INFO=$(ssh_run "echo \$(uname -s)_\$(uname -m)")
REMOTE_UNAME_S=$(echo "$REMOTE_INFO" | cut -d'_' -f1)
# -f2- 取第一个下划线之后的全部：x86_64 自带下划线，-f2 会截成 x86 被误判为不支持
REMOTE_UNAME_M=$(echo "$REMOTE_INFO" | cut -d'_' -f2-)

case "$REMOTE_UNAME_S" in
    Linux)   TARGET_OS="linux" ;;
    Darwin)  TARGET_OS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) TARGET_OS="windows" ;;
    *) echo -e "${RED}不支持的远程系统: $REMOTE_UNAME_S${NC}"; exit 1 ;;
esac

case "$REMOTE_UNAME_M" in
    x86_64|amd64)  TARGET_ARCH="amd64" ;;
    aarch64|arm64)  TARGET_ARCH="arm64" ;;
    armv7l)         TARGET_ARCH="arm" ;;
    *) echo -e "${RED}不支持的架构: $REMOTE_UNAME_M${NC}"; exit 1 ;;
esac

echo -e "${CYAN}==============================${NC}"
echo -e "${CYAN}   Admin 后台管理系统 - 部署   ${NC}"
echo -e "${CYAN}==============================${NC}"
echo ""
echo -e "  项目:       ${YELLOW}${PROJECT_NAME}${NC} (配置: $(basename "$ENV_FILE"))"
echo -e "  部署模式:   ${YELLOW}${DEPLOY_MODE}${NC}"
echo -e "  目标主机:   ${YELLOW}${SSH_USER}@${SSH_HOST}:${SSH_PORT}${NC}"
echo -e "  远程系统:   ${YELLOW}${TARGET_OS}/${TARGET_ARCH}${NC}"
if [ "$DEPLOY_MODE" != "admin" ]; then
echo -e "  后端目录:   ${YELLOW}${REMOTE_SERVER_DIR}${NC}"
fi
if [ "$DEPLOY_MODE" != "server" ]; then
echo -e "  前端目录:   ${YELLOW}${REMOTE_ADMIN_DIR}${NC}"
fi
echo ""

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# -------------------------------------------------------
# 编译后端
# -------------------------------------------------------
deploy_server() {
    echo -e "${YELLOW}[后端] 交叉编译 (${TARGET_OS}/${TARGET_ARCH})...${NC}"
    cd "$SERVER_DIR"

    SWAG_BIN="$(go env GOPATH)/bin/swag"
    if [ -f "$SWAG_BIN" ]; then
        echo -e "        生成 Swagger 文档..."
        "$SWAG_BIN" init -g main.go -o docs --parseDependency || true
    fi

    BIN_NAME="$SERVER_BIN_NAME"
    [ "$TARGET_OS" = "windows" ] && BIN_NAME="${BIN_NAME}.exe"

    CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -ldflags="-s -w" -o "$BUILD_DIR/$BIN_NAME" .
    cp config.prod.yaml "$BUILD_DIR/config.yaml"
    echo -e "${GREEN}        编译完成: $BIN_NAME (使用生产配置)${NC}"

    echo -e "${YELLOW}[后端] 上传到服务器...${NC}"
    check_remote_owner "$REMOTE_SERVER_DIR" "后端"
    ssh_run "mkdir -p ${REMOTE_SERVER_DIR}"
    scp_to "$BUILD_DIR/$BIN_NAME" "${REMOTE_SERVER_DIR}/$BIN_NAME"
    scp_to "$BUILD_DIR/config.yaml" "${REMOTE_SERVER_DIR}/config.yaml"
    ssh_run "chmod +x ${REMOTE_SERVER_DIR}/$BIN_NAME"
    mark_remote_owner "$REMOTE_SERVER_DIR"
    echo -e "${GREEN}        上传完成${NC}"

    if [ "$AUTO_RESTART" = "yes" ]; then
        ensure_systemd_service
        echo -e "${YELLOW}[后端] 重启服务 (${SERVICE_NAME})...${NC}"
        ssh_run "systemctl restart ${SERVICE_NAME}"
        echo -e "${GREEN}        服务已重启${NC}"
    fi
}

# -------------------------------------------------------
# 自动创建 systemd 服务
# -------------------------------------------------------
ensure_systemd_service() {
    local service_exists=0 unit_dir
    # 用 grep -q 的退出码判断存在性，不解析输出。
    # 不能写 `$(... | grep -c ...) || echo 0`：服务不存在时远程 grep -c 已输出 "0"
    # 且退出码为 1，|| 再补一个 0 会拼成两行 "0\n0"——既不等于 "0" 也不为空，
    # 结果跳过创建、直接 restart 报 service not found
    if ssh_run "systemctl list-unit-files ${SERVICE_NAME}.service 2>/dev/null | grep -q ${SERVICE_NAME}"; then
        service_exists=1
    fi

    # 服务已存在时校验归属：WorkingDirectory 指向别的目录说明服务名被其他项目占用
    if [ "$service_exists" = "1" ]; then
        unit_dir=$(ssh_run "systemctl show -p WorkingDirectory ${SERVICE_NAME} 2>/dev/null" | cut -d= -f2- | tr -d '[:space:]' || true)
        if [ -n "$unit_dir" ] && [ "$unit_dir" != "$REMOTE_SERVER_DIR" ]; then
            echo -e "${RED}systemd 服务 ${SERVICE_NAME} 已被其他项目使用 (WorkingDirectory=${unit_dir})${NC}"
            echo -e "${YELLOW}请在 $(basename "$ENV_FILE") 中为当前项目设置不同的 PROJECT_NAME 或 SERVICE_NAME${NC}"
            exit 1
        fi
    fi

    if [ "$service_exists" = "0" ]; then
        echo -e "${YELLOW}[后端] 创建 systemd 服务: ${SERVICE_NAME}...${NC}"
        ssh_run "cat > /etc/systemd/system/${SERVICE_NAME}.service << 'UNIT'
[Unit]
Description=${PROJECT_NAME} Server (${SERVICE_NAME})
After=network.target

[Service]
Type=simple
WorkingDirectory=${REMOTE_SERVER_DIR}
ExecStart=${REMOTE_SERVER_DIR}/${SERVER_BIN_NAME}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
UNIT"
        ssh_run "systemctl daemon-reload && systemctl enable ${SERVICE_NAME}"
        echo -e "${GREEN}        服务已创建并设为开机自启${NC}"
    fi
}

# -------------------------------------------------------
# 打包前端
# -------------------------------------------------------
deploy_admin() {
    echo -e "${YELLOW}[前端] 打包构建...${NC}"
    cd "$ADMIN_DIR"

    if [ ! -d node_modules ]; then
        echo -e "        安装依赖..."
        pnpm install --no-frozen-lockfile
    fi

    eval "$ADMIN_BUILD_CMD"

    DIST_DIR="$ADMIN_DIR/apps/web-antd/dist"
    if [ ! -d "$DIST_DIR" ]; then
        echo -e "${RED}前端打包输出目录不存在: $DIST_DIR${NC}"
        exit 1
    fi

    cd "$DIST_DIR"
    tar -czf "$BUILD_DIR/admin.tar.gz" .
    echo -e "${GREEN}        打包完成${NC}"

    echo -e "${YELLOW}[前端] 上传到服务器...${NC}"
    check_remote_owner "$REMOTE_ADMIN_DIR" "前端"
    ssh_run "mkdir -p ${REMOTE_ADMIN_DIR}"
    scp_to "$BUILD_DIR/admin.tar.gz" "$REMOTE_TMP_TAR"
    ssh_run "rm -rf ${REMOTE_ADMIN_DIR}/* && tar -xzf ${REMOTE_TMP_TAR} -C ${REMOTE_ADMIN_DIR} && rm -f ${REMOTE_TMP_TAR}"
    mark_remote_owner "$REMOTE_ADMIN_DIR"
    echo -e "${GREEN}        上传完成${NC}"
}

# -------------------------------------------------------
# 按模式执行
# -------------------------------------------------------
case "$DEPLOY_MODE" in
    server) deploy_server ;;
    admin)  deploy_admin ;;
    all)    deploy_server; deploy_admin ;;
esac

rm -rf "$BUILD_DIR"
rmdir "$ROOT_DIR/.build" 2>/dev/null || true

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   部署完成！(${PROJECT_NAME} / ${DEPLOY_MODE})${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
