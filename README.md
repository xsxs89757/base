# Admin 后台管理系统

基于 **Go Fiber + GORM + JWT** 后端 和 **Vben Admin (Vue 3 + Ant Design Vue)** 前端的后台管理基础框架。

## 技术栈

### 后端 (server/)
- **Fiber v2** - 高性能 Go Web 框架
- **GORM** - Go ORM 框架 (默认 SQLite，可切换 MySQL/PostgreSQL)
- **菜单权限码 RBAC** - 角色 → 菜单/按钮 auth_code → 路由，无额外策略表
- **JWT** - Token 认证
- **Swagger/OpenAPI** - API 文档自动生成

### 前端 (admin/)
- **Vue 3** + **TypeScript**
- **Vben Admin v5** (Ant Design Vue 版本)
- **Vite 7**
- **Pinia** 状态管理
- **Tailwind CSS**

## 快速开始

### 环境要求
- Go 1.24+
- Node.js 22+
- pnpm 10+

### 创建新项目

```bash
go run github.com/xsxs89757/base/tools/create-base@latest myproject
```

一条命令完成：保留共同 git 历史地克隆基底（之后才能 `make sync-base`）、生成随机
`jwt.secret`、按项目名填好 `.deploy.env` 与前端 `.env`、安装依赖、提交首个 commit。
常用参数：`--origin <仓库地址>`、`--title <前端标题>`、`--version main`、`--skip-install`。
手工创建方式见「基底与下游项目」。

### 一键启动（开发）

```bash
# 脚手架创建的项目已生成 server/config.yaml；手工克隆的需要先复制一份
cp server/config.yaml.example server/config.yaml

# 启动 (后端 air 热更新 + 前端 Vite HMR)
./dev.sh

# 或使用 Makefile 快捷命令
make dev

# 8080/5666 等开发端口被占用时，默认自动改用空闲端口启动，
# 同一台机器可同时跑多个项目的 dev，互不干扰。
# 如果希望杀死占用进程、坚持使用配置端口：
make dev-force
# 等价于
./dev.sh --force
```

后端修改 `.go` 文件后自动重新编译，前端修改即时热更新。按 `Ctrl+C` 停止所有服务。

### 公网调试（穿云内网穿透，可选）

微信/支付回调、手机真机联调、把本地页面发给别人看，都需要一个公网地址。装好
[穿云](https://github.com/xsxs89757/chuanyun) 桌面客户端并登录后，`./dev.sh` 会自动
把前后端各挂一条隧道，地址固定、重启不变：

```
  前端:    http://localhost:5666
  后端:    http://localhost:8080
  公网:    https://<用户>-<项目>-api.<域名>     (穿云 -> 后端)
  ➜  穿云:  https://<用户>-<项目>-admin.<域名>  (前端 dev server)
```

- **穿云没装/没开/没登录时全部静默跳过**，本地开发照常，不会因此启动失败；
- 隧道名按项目区分（取 `.deploy.env` 的 `PROJECT_NAME`，没有则用仓库目录名），
  同一台机器跑多个项目互不抢名字；
- 端口自动避让后隧道会跟着指向真实端口，上次残留的隧道会先清掉再重建；
- 退出时两条隧道都会注销；
- 后端进程可读环境变量 `CHUANYUN_PUBLIC_URL` 拼回调地址（未接入时为空，业务代码自行回落本地）；
- 本次不想用：`./dev.sh --no-chuanyun`（等价 `CHUANYUN=0 ./dev.sh`）。

#### 用 `chuanyun.toml` 声明额外隧道 / 接入同事的服务

后端和 admin 前端不用配，dev.sh 已经自动处理。需要**多暴露一个固定端口**，或者
**把同事的服务接到本机**时，复制模板即可（`cp chuanyun.toml.example chuanyun.toml`）：

```toml
project = "base"          # 隧道名前缀，决定公网地址

[[tunnels]]               # 我暴露的：dev.sh 不启动的服务（手工进程 / docker）
name = "docs"             # → {用户}-base-docs.{域名}
port = 9000               # 按原样使用，不做端口避让

[[connects]]              # 我接入的：后端用同事的
local_port = 8082
from = "zhangsan-api"     # 同事的隧道子域名，也可写完整 URL
```

对方隧道设了访问口令的话，口令写在 **`chuanyun.local.toml`**（已 gitignore，每人自己一份），
按 `from` 对上即可：

```toml
[[connects]]
from = "zhangsan-api"
auth = "user:pass"
```

`chuanyun.toml` 进 git，口令一旦写进去就永远留在历史里，所以分成两个文件。

`./dev.sh` 启动时自动应用，退出时把自己建的隧道**关掉**（不删——你在客户端给它设的口令留着，
下次启动原地打开）；手工建的隧道不会被动到。

**自己隧道的口令**不在这两个文件里配：那是隧道主人的事，不是项目的事。在穿云客户端
新建隧道时填「访问口令」那一栏，dev.sh 每次重新注册时会保留它。

**为什么必须有 `project` 前缀**：公网地址的构成是 `{用户}-{name}.{域名}`，
里面**不含项目信息**——两个项目都写 `name = "api"`，同一个人名下就会撞名，
后启动的那个会被拒绝。加上 project 前缀（`base-api` / `crm-api`）才互不干扰。
不写 `project` 时回落到 `.deploy.env` 的 `PROJECT_NAME`，再没有就用仓库目录名。

**`[[tunnels]]` 别写后端/前端的端口**：dev.sh 在端口被占时会自动改用空闲端口，
而 toml 里是写死的，两者会不一致。dev.sh 自己启动的服务一律由它动态建隧道；
`dev.project.sh` 里的服务用 `chuanyun_up` 现取现用（见「新增额外服务」）。

**Windows 用户**：在 **Git Bash** 中运行 `./dev.sh`（随 Git for Windows 附带，勿用
PowerShell/cmd）。脚本会自动切换到 Windows 实现——air 改用 `server/.air.windows.toml`
（无 Unix 内联环境变量前缀、产物带 `.exe`）、端口探测改用 `netstat`、结束进程改用
`taskkill`，其余用法与 macOS/Linux 完全一致。

### 手动启动

**后端：**
```bash
cd server
cp config.yaml.example config.yaml  # 首次
go mod tidy
go run main.go
```
后端运行在 `http://localhost:8080`

**前端：**
```bash
cd admin
pnpm install
pnpm dev:antd
```
前端运行在 `http://localhost:5666`

### 默认账号

| 账号 | 密码 | 角色 | 说明 |
|------|------|------|------|
| super | 123456 | super | 超级管理员，拥有所有权限 |
| admin | 123456 | admin | 管理员 |
| jack | 123456 | user | 普通用户，仅查看权限 |

> 超级管理员 (id=1) 不受任何权限限制，不会出现在用户管理列表中，不可被修改或删除。

## 项目结构

```
├── Makefile                     # 快捷命令入口
├── CHANGELOG.md                 # 版本记录与升级步骤
├── .base-version                # 当前基底版本 (由基底发布流程写入，下游勿改)
├── dev.sh                       # 一键开发启动 (air 热更新)
├── deploy.sh                    # 一键部署脚本
├── .deploy.env.example          # 部署配置模板
├── .github/workflows/ci.yml     # CI: 挂载点校验 / 后端测试 / 前端构建
├── scripts/check-hooks.sh       # 校验基底未改动下游挂载点
├── tools/create-base/           # 新项目脚手架 (独立 Go module)
│
├── server/                      # Go 后端
│   ├── main.go                  # 入口文件
│   ├── config.yaml.example      # 配置模板
│   ├── .air.toml                # air 热更新配置 (macOS/Linux)
│   ├── .air.windows.toml        # air 热更新配置 (Windows，dev.sh 自动选用)
│   ├── config/                  # 配置解析
│   ├── docs/                    # Swagger 生成物 (随仓库提交，main.go 依赖它编译)
│   └── internal/
│       ├── dto/                 # 数据传输对象
│       │   ├── admin/           # 后台管理 DTO
│       │   └── base.go          # 通用响应结构
│       ├── handler/             # 路由处理器 (含 Swagger 注解)
│       │   ├── admin/           # 后台管理 API
│       │   └── api/             # 前台 API (预留)
│       ├── middleware/          # JWT / 权限码 / 操作日志中间件
│       ├── model/               # GORM 数据模型
│       │   └── admin/           # 后台管理模型
│       ├── router/              # 路由定义
│       ├── service/             # 业务逻辑
│       │   └── admin/           # 后台管理服务
│       ├── store/               # 数据库初始化 & 种子数据
│       └── validator/           # 请求验证
│
└── admin/                       # Vben Admin 后台前端
    ├── apps/
    │   └── web-antd/            # Ant Design Vue 应用
    │       └── src/
    │           ├── adapter/     # 组件适配器
    │           ├── api/         # API 接口定义
    │           ├── locales/     # 国际化 (中/英)
    │           └── views/       # 页面
    │               └── system/  # 系统管理模块
    └── packages/                # 共享包
```

## 部署

### 配置

```bash
cp .deploy.env.example .deploy.env
# 编辑 .deploy.env 填入服务器 SSH 信息和目录
```

### 部署命令

```bash
./deploy.sh all      # 全量部署 (默认)
./deploy.sh server   # 仅部署后端
./deploy.sh admin    # 仅部署后台前端

# Makefile 快捷命令
make release          # 全量部署
make release-server   # 仅部署后端
make release-admin    # 仅部署后台前端
```

### 多项目发布到同一台服务器

base 的典型用法是**每个项目一份仓库拷贝**。每个项目在自己的 `.deploy.env` 里
设置独立的 `PROJECT_NAME` 和远程目录，即可安全地发布到同一台服务器：

```bash
# 项目 A 仓库的 .deploy.env          # 项目 B 仓库的 .deploy.env
PROJECT_NAME=shop                    PROJECT_NAME=blog
REMOTE_SERVER_DIR=/opt/shop/server   REMOTE_SERVER_DIR=/opt/blog/server
REMOTE_ADMIN_DIR=/opt/shop/admin     REMOTE_ADMIN_DIR=/opt/blog/admin
```

部署按 `PROJECT_NAME` 隔离：systemd 服务名（`shop-server` / `blog-server`）、
远程临时包和本地构建目录互不影响。同时有三道防覆盖保护，任一触发都会中止部署：

- 缺少项目标识（`PROJECT_NAME` / `SERVICE_NAME` 都未设置）时拒绝部署；
- 远程目录内有 `.deploy-project` 归属标记，目录属于其他项目时报错；
- systemd 服务名已被指向其他目录的项目占用时报错。

如果想在**同一份仓库里维护多套发布目标**（比如一套代码发多个站点、或区分测试/生产），
再用 `.deploy.<名字>.env` 配置：

```bash
cp .deploy.env.example .deploy.shop.env   # 编辑填入该目标的项目名/目录

./deploy.sh all shop         # 使用 .deploy.shop.env 全量部署
./deploy.sh server -p shop   # 仅部署 shop 后端
./deploy.sh --list           # 查看已有的部署配置

# Makefile 等价命令
make release PROJECT=shop
make release-server PROJECT=shop
```

不指定项目名时仍读取 `.deploy.env`，与单项目用法完全兼容。

部署脚本会：
- **自动检测**远程服务器的系统和架构 (linux/amd64, linux/arm64 等)
- **端口自动避让**（仅首次部署）：配置端口在服务器上被占用时，从该端口向后找到
  空闲端口启动，并**同步回写本地 `server/config.prod.yaml`**；终端会提示 nginx
  站点配置应使用的实际端口
- **端口冲突保护**（已部署过的服务）：端口被其他进程占用时**只告警不自动更换**
  （nginx/回调地址都依赖既定端口），需人工释放端口或改配置后重新部署
- **交叉编译** Go 后端 (`CGO_ENABLED=0`)
- **打包** 前端静态资源
- **SSH 上传**到服务器指定目录
- **自动创建** systemd 服务（首次部署时）
- **自动重启**服务并做健康检查，启动失败自动拉取最近的 journal 日志

## 新增额外服务（下游）

在 server/admin 之外再加一个服务（Nuxt 前台 `web/`、任务进程、回调网关等），
**不要改 `dev.sh` / `deploy.sh` 本体**——用仓库根的两个挂载点，`make sync-base` 永不冲突。

### 本地开发：`dev.project.sh`

新建这个文件即可（基底不含它，也承诺永不创建）。以加一个 Nuxt 前台 `web/` 为例：

```bash
#!/bin/bash
# dev.project.sh —— 被 ./dev.sh 自动 source
WEB_PORT="${WEB_PORT:-3000}"
WEB_PID=""
WEB_URL=""

project_dev_start() {
    # 1) 解析端口：自动避开占用、遵守 --force、不与本次其他服务撞port
    resolve_port "$WEB_PORT" "网站"
    WEB_PORT="$RESOLVED_PORT"

    echo -e "${YELLOW}[+] 启动网站 (http://localhost:${WEB_PORT})${NC}"
    cd "$ROOT_DIR/web"
    [ -d node_modules ] || pnpm install

    # 2) 端口用环境变量传给子进程（`-- --port` 会被 pnpm 吞掉）
    #    后端地址用已解析的 SERVER_PORT
    WEB_PORT="$WEB_PORT" NUXT_API_BASE="http://localhost:${SERVER_PORT}/api" pnpm dev &
    WEB_PID=$!
    sleep 3

    # 3) 要公网地址就建条隧道；穿云没开时 CHUANYUN_LAST_URL 是空字符串
    chuanyun_up "$(chuanyun_slug)-web" "$WEB_PORT"
    WEB_URL="$CHUANYUN_LAST_URL"
}

project_dev_stop() {
    # pnpm/nuxt 是多层包装，必须杀进程树
    [ -n "$WEB_PID" ] && kill_tree "$WEB_PID" && echo -e "${GREEN}网站已停止${NC}"
    return 0
}

project_dev_info() {
    echo -e "  网站:    ${CYAN}http://localhost:${WEB_PORT}${NC}"
    [ -n "$WEB_URL" ] && echo -e "  网站公网: ${CYAN}${WEB_URL}${NC}"
    return 0
}
```

`./dev.sh` 的输出就会多出两行：

```
  网站:    http://localhost:3000
  网站公网: https://<用户>-<项目>-web.<域名>
```

可直接用的助手：

| 助手 | 作用 |
|---|---|
| `resolve_port <端口> <名称>` | 解析端口，结果在 `RESOLVED_PORT`。**别写死端口**，否则同机多项目互抢 |
| `chuanyun_up <名字> <端口>` | 建公网隧道，地址在 `CHUANYUN_LAST_URL`；退出时 dev.sh 统一注销 |
| `chuanyun_slug` | 项目标识（`.deploy.env` 的 `PROJECT_NAME`，没有则仓库目录名） |
| `kill_tree <PID>` | 结束进程树，`project_dev_stop` 里用 |
| `ROOT_DIR` / `SERVER_PORT` | 仓库根 / 已解析的后端端口 |

> **注意 dev server 的 Host 校验**：Vite 系（含 Nuxt 的 vite 层）会拒绝陌生 Host，
> 经隧道访问报 `Blocked request`。后台 admin 由 `vite-plugin-chuanyun` 自动处理，
> 自己加的服务需要把隧道域名加进该框架的 `allowedHosts`。

### 部署：`deploy.project.sh`

```bash
#!/bin/bash
# deploy.project.sh —— 被 ./deploy.sh 自动 source
PROJECT_DEPLOY_TARGETS="web"

project_deploy_web() {
    echo -e "${YELLOW}[网站] 打包构建...${NC}"
    cd "$ROOT_DIR/web" && pnpm install && pnpm build

    check_remote_owner "$REMOTE_WEB_DIR" "网站"
    ssh_run "mkdir -p ${REMOTE_WEB_DIR}"
    scp_to "..." "${REMOTE_WEB_DIR}/..."
    mark_remote_owner "$REMOTE_WEB_DIR"
}
```

之后 `./deploy.sh web` 单独发布，`./deploy.sh all` 会在 server/admin 之后一并执行；
`make release-web` 同样可用。所需变量（如 `REMOTE_WEB_DIR`）加进 `.deploy.env` 即可。

## 基底与下游项目

本仓库是统一基底：`https://github.com/xsxs89757/base`。新项目从基底克隆派生，
之后随时 `make sync-base` 合入基底的 bug 修复和新功能。

```bash
# 创建新项目：推荐用脚手架，它会做完下面所有事情
go run github.com/xsxs89757/base/tools/create-base@latest myproject --origin <新项目仓库地址>

# 手工方式（等价，必须保留共同 git 历史，禁止删 .git 重新 init / 纯文件拷贝）
git clone --no-tags -o base https://github.com/xsxs89757/base.git myproject
cd myproject
git remote add origin <新项目仓库地址>
git branch --unset-upstream        # 否则 git push 会推向基底
git push -u origin main

# 之后同步基底更新
make sync-base                     # 默认合入基底最新版本标签
make sync-base VERSION=v1.2.0      # 合入指定版本
make sync-base VERSION=main        # 合入开发中的 main
make base-version                  # 查看当前已合入版本与远端最新版本
```

`--no-tags` 不能省：基底的 `v*` 标签直接拉进来会和下游自己的版本号撞车。
`make sync-base` 会把基底标签拉到 `base/v*` 命名空间，与下游标签隔离；
下游用 `git describe` 时记得加 `--match 'v*'` 排除掉它们。

下游开发约定（按 sync-base 合并成本分两类）：

- **核心框架不建议就地改**（server 框架层、admin 的 vben 封装）：这些文件基底
  会持续更新，下游改了每次同步都要重复解决冲突——通用改进请回流基底仓库，
  改完各下游 `make sync-base` 合入；
- **其余自由改**：CLAUDE.md / README / dev.sh / deploy.sh / Makefile 等脚手架
  和文档尽管项目化；业务代码放新增文件；路由和模型用两个专属挂载点注册
  （基底永不改动它们）：`server/internal/router/project.go`、
  `server/internal/store/project.go`；
- 唯一硬性禁令：不改 `server/go.mod` 的 **module 名**（保持 `base`），否则 import
  路径全面 diverge，之后每次 merge 大面积冲突；**新增依赖不受限**——`go get`
  照常用，同步冲突时合并双方依赖行后 `go mod tidy` 即可。

详细纪律见 CLAUDE.md / AGENTS.md 的「基底与下游项目」一节。

## 功能模块

| 模块 | 说明 |
|------|------|
| 用户管理 | 用户增删改查、角色分配、状态切换 |
| 角色管理 | 角色增删改查、菜单权限分配 |
| 菜单管理 | 菜单/目录/按钮管理、树形结构 |
| 部门管理 | 部门树形管理 |
| 配置管理 | 系统参数配置、按分组筛选 |
| 操作日志 | 自动记录 POST/PUT/DELETE 操作 |

## API 文档

### Swagger UI
启动后端后访问：**http://localhost:8080/swagger/index.html**

### OpenAPI 导入
可将以下文件导入 Postman、Apifox、YApi 等 API 管理工具：
- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: `server/docs/swagger.yaml`

> 生产环境中 Swagger 默认关闭，通过 `config.yaml` 中 `enable_swagger: true` 开启。

## API 接口

### 认证
- `POST /admin/auth/login` - 登录
- `POST /admin/auth/logout` - 登出
- `POST /admin/auth/refresh` - 刷新 Token
- `POST /admin/auth/change-password` - 修改密码
- `GET /admin/auth/codes` - 获取权限码

### 用户
- `GET /admin/user/info` - 当前用户信息

### 菜单
- `GET /admin/menu/all` - 获取用户菜单 (前端路由)

### 系统管理
- `GET/POST/PUT/DELETE /admin/system/user/*` - 用户管理
- `GET/POST/PUT/DELETE /admin/system/role/*` - 角色管理
- `GET/POST/PUT/DELETE /admin/system/menu/*` - 菜单管理
- `GET/POST/PUT/DELETE /admin/system/dept/*` - 部门管理
- `GET/POST/PUT/DELETE /admin/system/config/*` - 配置管理
- `GET/DELETE /admin/system/operation-log/*` - 操作日志

> 前端通过 Vite 代理 `/api` → `/admin`

## 权限说明

权限模型是"菜单权限码 RBAC"，不依赖额外的策略表：

- 每个需要保护的接口在 `server/internal/middleware/permission.go` 的路由表里映射到一个权限码（如 `System:User:Edit`），权限码存在菜单/按钮的 `auth_code` 字段上；
- 角色通过 `role_menus` 关联菜单，用户持有的**启用**角色中任一关联了该权限码的启用菜单即放行；未登记的 `/admin` 路由对非 super 一律 403；
- `JWTAuth` 每个请求以数据库为准核对用户状态和角色（进程内缓存 1 分钟，用户/角色变更即时失效），禁用用户、调整角色、修改密码立即生效，不用等 token 过期；
- **super** 角色和 id=1 的内置超管绕过全部权限判定，且不能被普通管理员修改/删除；
- **admin** 角色：种子默认拥有系统管理全部菜单；**user** 角色：仅基础查看。

下游项目给自己的路由登记权限码（写在 `server/internal/router/project.go`）：

```go
middleware.RegisterRoutePermissions(
    middleware.RoutePermission{Method: "GET", Path: "/admin/shop/order/list", Code: "Shop:Order:List"},
    middleware.RoutePermission{Method: "PUT", Path: "/admin/shop/order/:id", Code: "Shop:Order:Edit"},
)
```

只需登录、不校验权限码的路由用 `middleware.RegisterAuthenticatedRoutes`。`middleware.CasbinAuth()` 是 `PermissionAuth()` 的旧名，仍可使用。

## 版本与发布

基底按语义化版本发布，每个版本的改动和升级步骤记在 [CHANGELOG.md](CHANGELOG.md)。
`.base-version` 记录当前代码来自哪个基底版本（由基底发布流程写入，下游不要手改）。

版本号含义：

- **MAJOR**：同步后需要 merge + `go mod tidy` 之外的人工迁移（删配置键、改挂载点签名、Vben 大版本）；
- **MINOR**：新功能、新增可选配置或菜单，可能需要用户重新登录；
- **PATCH**：修 bug、改文档。

基底维护者发布新版本（下游用不到）：

```bash
# 1. 先把内容提交推到 main，等 CI 通过
# 2. 在 CHANGELOG.md 写好 ## [1.1.0] - YYYY-MM-DD 条目
make base-release VERSION=v1.1.0
```

`base-release` 会先跑 `make base-check`（挂载点冻结校验、后端测试、交叉编译、脚本语法、
Swagger 文档时效），再写 `.base-version`、打标签、原子推送。

## 切换数据库

修改 `server/config.yaml` 中的数据库配置：

```yaml
# MySQL
database:
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=True&loc=Local"

# PostgreSQL
database:
  driver: postgres
  dsn: "host=localhost user=postgres password=postgres dbname=admin port=5432 sslmode=disable"

# SQL Server
database:
  driver: sqlserver
  dsn: "sqlserver://user:password@localhost:1433?database=admin"
```

项目已内置 SQLite、MySQL/MariaDB、PostgreSQL、SQL Server 对应的 GORM 驱动，无需额外 `go get`。

### 本地 SQLite 与生产 MySQL 的差异陷阱

本地用 SQLite 开发、生产切 MySQL 时，有一类问题只会在 MySQL 上暴露，本地测不出来：

- **同一字段禁止同时写 `uniqueIndex` 和 `index`**，例如
  `gorm:"uniqueIndex;index"`。两个未命名标签会生成同名的默认索引，
  GORM 把同一列在一个索引里放两次——本地 SQLite 建表不报错，
  MySQL AutoMigrate 直接报 `1060 Duplicate column name`。
  `uniqueIndex` 本身就是索引，不需要再叠加 `index`。
- 上线前务必用 MySQL 完整启动一次（AutoMigrate + 主流程），
  不要只依赖本地 SQLite 验证。
