# Admin 后台管理系统

基于 **Go Fiber + GORM + Casbin + JWT** 后端 和 **Vben Admin (Vue 3 + Ant Design Vue)** 前端的后台管理基础框架。

## 技术栈

### 后端 (server/)
- **Fiber v2** - 高性能 Go Web 框架
- **GORM** - Go ORM 框架 (默认 SQLite，可切换 MySQL/PostgreSQL)
- **Casbin v3** - 基于 RBAC 的权限控制
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

### 一键启动（开发）

```bash
# 首次使用需复制配置文件
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
├── dev.sh                       # 一键开发启动 (air 热更新)
├── deploy.sh                    # 一键部署脚本
├── .deploy.env.example          # 部署配置模板
│
├── server/                      # Go 后端
│   ├── main.go                  # 入口文件
│   ├── config.yaml.example      # 配置模板
│   ├── .air.toml                # air 热更新配置
│   ├── config/                  # 配置解析
│   ├── docs/                    # Swagger 自动生成文档
│   └── internal/
│       ├── dto/                 # 数据传输对象
│       │   ├── admin/           # 后台管理 DTO
│       │   └── base.go          # 通用响应结构
│       ├── handler/             # 路由处理器 (含 Swagger 注解)
│       │   ├── admin/           # 后台管理 API
│       │   └── api/             # 前台 API (预留)
│       ├── middleware/          # JWT / Casbin / 操作日志中间件
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

## 基底与下游项目

本仓库是统一基底：`https://github.com/xsxs89757/base`。新项目从基底克隆派生，
之后随时 `make sync-base` 合入基底的 bug 修复和新功能。

```bash
# 创建新项目（必须保留共同 git 历史，禁止删 .git 重新 init / 纯文件拷贝）
git clone https://github.com/xsxs89757/base.git myproject
cd myproject
git remote rename origin base
git remote add origin <新项目仓库地址>
git push -u origin main

# 之后同步基底更新
make sync-base   # 等价 git fetch base && git merge base/main
```

下游开发约定（保证 merge 基本无冲突）：

- 基底文件只在基底仓库改，改完在各下游 `make sync-base` 合入；
- 业务代码放新增文件；路由和模型用两个专属挂载点注册（基底永不改动它们）：
  `server/internal/router/project.go`、`server/internal/store/project.go`；
- 不改 `server/go.mod` 的 module 名（保持 `base`）。

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

系统采用 Casbin RBAC 模型：
- **super** 角色：超级管理员，Casbin 和菜单/权限码全部绕过
- **admin** 角色：可读写系统管理模块
- **user** 角色：仅可查看基础信息

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
