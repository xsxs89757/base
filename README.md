# Admin 后台管理系统

基于 **Go Fiber + GORM + Casbin + JWT** 后端 和 **Vben Admin (Vue 3 + Ant Design Vue)** 前端的后台管理基础框架。

## 技术栈

### 后端 (server/)
- **Fiber v2** - 高性能 Go Web 框架
- **GORM** - Go ORM 框架 (默认 SQLite，可切换 MySQL/PostgreSQL)
- **Casbin v3** - 基于 RBAC 的权限控制
- **JWT** - Token 认证
- **Swagger/OpenAPI** - API 文档自动生成
- **SQLite** - 开发环境数据库

### 前端 (web/)
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

### 一键启动

```bash
./dev.sh
```

自动编译后端、生成 Swagger 文档、启动前后端服务。按 `Ctrl+C` 停止所有服务。

### 手动启动

**后端：**
```bash
cd server
go mod tidy
go run main.go
```
后端运行在 `http://localhost:8080`

**前端：**
```bash
cd web
pnpm install
pnpm dev:antd
```
前端运行在 `http://localhost:5666`

### 默认账号

| 账号 | 密码 | 角色 | 说明 |
|------|------|------|------|
| vben | 123456 | super | 超级管理员，拥有所有权限 |
| admin | 123456 | admin | 管理员 |
| jack | 123456 | user | 普通用户，仅查看权限 |

## 项目结构

```
├── dev.sh                     # 一键启动脚本
├── server/                    # Go 后端
│   ├── main.go                # 入口文件
│   ├── config.yaml            # 配置文件
│   ├── config/                # 配置解析
│   ├── docs/                  # Swagger 自动生成文档
│   │   ├── docs.go
│   │   ├── swagger.json       # OpenAPI JSON (可导入 Postman/Apifox)
│   │   └── swagger.yaml       # OpenAPI YAML
│   ├── internal/
│   │   ├── database/          # 数据库初始化 & 种子数据
│   │   ├── handler/           # 路由处理器 (含 Swagger 注解)
│   │   ├── middleware/        # JWT 认证 & Casbin 权限中间件
│   │   ├── model/             # GORM 数据模型
│   │   ├── router/            # 路由定义
│   │   └── service/           # 业务逻辑
│   └── rbac/                  # Casbin RBAC 配置
│
├── web/                       # Vben Admin 前端
│   ├── apps/
│   │   ├── web-antd/          # Ant Design Vue 应用
│   │   └── backend-mock/      # Mock 后端 (开发可选)
│   ├── packages/              # 共享包
│   └── internal/              # 内部构建工具
│
└── README.md
```

## API 文档

### Swagger UI
启动后端后访问：**http://localhost:8080/swagger/index.html**

### OpenAPI 导入
可将以下文件导入 Postman、Apifox、YApi 等 API 管理工具：
- **JSON**: `http://localhost:8080/swagger/doc.json` 或 `server/docs/swagger.json`
- **YAML**: `server/docs/swagger.yaml`

### 重新生成文档
修改 handler 中的 Swagger 注解后，运行：
```bash
cd server
# 安装 swag CLI (仅首次)
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
$(go env GOPATH)/bin/swag init -g main.go -o docs --parseDependency --parseInternal
```

## API 接口

### 认证
- `POST /api/auth/login` - 登录
- `POST /api/auth/logout` - 登出
- `POST /api/auth/refresh` - 刷新 Token
- `GET /api/auth/codes` - 获取权限码

### 用户
- `GET /api/user/info` - 当前用户信息

### 菜单
- `GET /api/menu/all` - 获取用户菜单 (前端路由)

### 系统管理
- `GET/POST/PUT/DELETE /api/system/role/*` - 角色管理
- `GET/POST/PUT/DELETE /api/system/menu/*` - 菜单管理
- `GET/POST/PUT/DELETE /api/system/dept/*` - 部门管理
- `GET/POST/PUT/DELETE /api/system/user/*` - 用户管理

## 权限说明

系统采用 Casbin RBAC 模型：
- **super** 角色：超级管理员，自动拥有所有权限
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
```

需要同时安装对应的 GORM 驱动：
```bash
go get gorm.io/driver/mysql
# 或
go get gorm.io/driver/postgres
```
