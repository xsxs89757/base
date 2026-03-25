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
BUILD_DIR="$ROOT_DIR/.build"

# -------------------------------------------------------
# 使用说明
# -------------------------------------------------------
usage() {
    echo -e "${CYAN}用法: ./deploy.sh [server|web|all]${NC}"
    echo ""
    echo "  server  - 仅部署后端"
    echo "  web     - 仅部署前端"
    echo "  all     - 全量部署 (默认)"
    exit 0
}

DEPLOY_MODE="${1:-all}"
case "$DEPLOY_MODE" in
    server|web|all) ;;
    -h|--help|help) usage ;;
    *) echo -e "${RED}未知模式: $DEPLOY_MODE${NC}"; usage ;;
esac

# -------------------------------------------------------
# 加载配置
# -------------------------------------------------------
ENV_FILE="$ROOT_DIR/.deploy.env"
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}缺少部署配置文件 .deploy.env${NC}"
    echo -e "${YELLOW}请复制 .deploy.env.example 为 .deploy.env 并填写实际配置${NC}"
    echo ""
    echo "  cp .deploy.env.example .deploy.env"
    exit 1
fi

source "$ENV_FILE"

for var in SSH_HOST SSH_USER SSH_PASS REMOTE_SERVER_DIR REMOTE_WEB_DIR; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}配置项 $var 未设置，请检查 .deploy.env${NC}"
        exit 1
    fi
done

SSH_PORT="${SSH_PORT:-22}"
AUTO_RESTART="${AUTO_RESTART:-no}"
SERVICE_NAME="${SERVICE_NAME:-admin-server}"
SERVER_BIN_NAME="${SERVER_BIN_NAME:-server}"
WEB_BUILD_CMD="${WEB_BUILD_CMD:-pnpm build:antd}"

# 检测 sshpass
if ! command -v sshpass &>/dev/null; then
    echo -e "${YELLOW}安装 sshpass...${NC}"
    if [[ "$(uname)" == "Darwin" ]]; then
        brew install hudochenkov/sshpass/sshpass
    else
        sudo apt-get install -y sshpass 2>/dev/null || sudo yum install -y sshpass 2>/dev/null
    fi
fi

SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10 -p ${SSH_PORT}"
ssh_run() { sshpass -p "$SSH_PASS" ssh $SSH_OPTS "${SSH_USER}@${SSH_HOST}" "$@"; }
scp_to() { sshpass -p "$SSH_PASS" scp $SSH_OPTS -P "${SSH_PORT}" "$1" "${SSH_USER}@${SSH_HOST}:$2"; }

# -------------------------------------------------------
# 自动检测远程系统和架构
# -------------------------------------------------------
echo -e "${YELLOW}检测远程服务器系统信息...${NC}"
REMOTE_INFO=$(ssh_run "echo \$(uname -s)_\$(uname -m)")
REMOTE_UNAME_S=$(echo "$REMOTE_INFO" | cut -d'_' -f1)
REMOTE_UNAME_M=$(echo "$REMOTE_INFO" | cut -d'_' -f2)

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
echo -e "  部署模式:   ${YELLOW}${DEPLOY_MODE}${NC}"
echo -e "  目标主机:   ${YELLOW}${SSH_USER}@${SSH_HOST}:${SSH_PORT}${NC}"
echo -e "  远程系统:   ${YELLOW}${TARGET_OS}/${TARGET_ARCH}${NC}"
if [ "$DEPLOY_MODE" != "web" ]; then
echo -e "  后端目录:   ${YELLOW}${REMOTE_SERVER_DIR}${NC}"
fi
if [ "$DEPLOY_MODE" != "server" ]; then
echo -e "  前端目录:   ${YELLOW}${REMOTE_WEB_DIR}${NC}"
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
        "$SWAG_BIN" init -g main.go -o docs --parseDependency --parseInternal 2>/dev/null || true
    fi

    BIN_NAME="$SERVER_BIN_NAME"
    [ "$TARGET_OS" = "windows" ] && BIN_NAME="${BIN_NAME}.exe"

    CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -ldflags="-s -w" -o "$BUILD_DIR/$BIN_NAME" .
    cp config.prod.yaml "$BUILD_DIR/config.yaml"
    echo -e "${GREEN}        编译完成: $BIN_NAME (使用生产配置)${NC}"

    echo -e "${YELLOW}[后端] 上传到服务器...${NC}"
    ssh_run "mkdir -p ${REMOTE_SERVER_DIR}"
    scp_to "$BUILD_DIR/$BIN_NAME" "${REMOTE_SERVER_DIR}/$BIN_NAME"
    scp_to "$BUILD_DIR/config.yaml" "${REMOTE_SERVER_DIR}/config.yaml"
    ssh_run "chmod +x ${REMOTE_SERVER_DIR}/$BIN_NAME"
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
    local service_exists
    service_exists=$(ssh_run "systemctl list-unit-files ${SERVICE_NAME}.service 2>/dev/null | grep -c ${SERVICE_NAME}" || echo "0")

    if [ "$service_exists" = "0" ]; then
        echo -e "${YELLOW}[后端] 创建 systemd 服务: ${SERVICE_NAME}...${NC}"
        ssh_run "cat > /etc/systemd/system/${SERVICE_NAME}.service << 'UNIT'
[Unit]
Description=Admin Server
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
deploy_web() {
    echo -e "${YELLOW}[前端] 打包构建...${NC}"
    cd "$WEB_DIR"

    if [ ! -d node_modules ]; then
        echo -e "        安装依赖..."
        pnpm install --no-frozen-lockfile
    fi

    eval "$WEB_BUILD_CMD"

    DIST_DIR="$WEB_DIR/apps/web-antd/dist"
    if [ ! -d "$DIST_DIR" ]; then
        echo -e "${RED}前端打包输出目录不存在: $DIST_DIR${NC}"
        exit 1
    fi

    cd "$DIST_DIR"
    tar -czf "$BUILD_DIR/web.tar.gz" .
    echo -e "${GREEN}        打包完成${NC}"

    echo -e "${YELLOW}[前端] 上传到服务器...${NC}"
    ssh_run "mkdir -p ${REMOTE_WEB_DIR}"
    scp_to "$BUILD_DIR/web.tar.gz" "/tmp/web.tar.gz"
    ssh_run "rm -rf ${REMOTE_WEB_DIR}/* && tar -xzf /tmp/web.tar.gz -C ${REMOTE_WEB_DIR} && rm -f /tmp/web.tar.gz"
    echo -e "${GREEN}        上传完成${NC}"
}

# -------------------------------------------------------
# 按模式执行
# -------------------------------------------------------
case "$DEPLOY_MODE" in
    server) deploy_server ;;
    web)    deploy_web ;;
    all)    deploy_server; deploy_web ;;
esac

rm -rf "$BUILD_DIR"

echo ""
echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}   部署完成！(${DEPLOY_MODE})${NC}"
echo -e "${GREEN}==============================${NC}"
echo ""
