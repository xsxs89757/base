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
```

部署脚本会：
- **自动检测**远程服务器的系统和架构 (linux/amd64, linux/arm64 等)
- **交叉编译** Go 后端 (`CGO_ENABLED=0`)
- **打包** 前端静态资源
- **SSH 上传**到服务器指定目录
- **自动创建** systemd 服务（首次部署时）
- **自动重启**服务

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
