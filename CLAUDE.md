# Claude Code 项目规则

本文档定义了在此项目中使用 Claude Code 时应遵循的规则和约定。

## 项目概述

这是一个基于 **Go Fiber + GORM + Casbin + JWT** 后端和 **Vben Admin (Vue 3 + Ant Design Vue)** 前端的全栈后台管理系统。

## 技术栈

### 后端
- **框架**: Go Fiber v2
- **ORM**: GORM (默认 SQLite)
- **权限**: Casbin v3 (RBAC)
- **认证**: JWT
- **文档**: Swagger/OpenAPI
- **热重载**: Air

### 前端
- **框架**: Vue 3 + TypeScript
- **UI**: Vben Admin v5 (Ant Design Vue)
- **构建**: Vite 7
- **状态**: Pinia
- **样式**: Tailwind CSS

## 代码规范

### Go 后端规范

1. **项目结构**
   - `internal/handler/` - HTTP 处理器，包含 Swagger 注解
   - `internal/service/` - 业务逻辑层
   - `internal/model/` - GORM 数据模型
   - `internal/dto/` - 数据传输对象
   - `internal/validator/` - 请求验证
   - `internal/middleware/` - 中间件 (JWT, Casbin, 日志)
   - `internal/router/` - 路由定义

2. **命名约定**
   - 文件名使用小写下划线: `operation_log.go`
   - 包名使用小写单词: `admin`, `handler`
   - 结构体使用大驼峰: `UserInfo`, `LoginRequest`
   - 方法使用大驼峰: `GetUserInfo()`, `CreateRole()`

3. **API 开发流程**
   - 在 `model/` 定义数据模型
   - 在 `dto/` 定义请求/响应结构
   - 在 `validator/` 添加验证规则
   - 在 `service/` 实现业务逻辑
   - 在 `handler/` 添加 HTTP 处理器和 Swagger 注解
   - 在 `router/` 注册路由
   - 运行 `swag init` 更新 API 文档

4. **Swagger 注解**
   - 每个 handler 方法必须包含完整的 Swagger 注解
   - 包括: `@Summary`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`
   - 修改后必须重新生成文档: `swag init -g main.go -o docs --parseDependency --parseInternal`

5. **错误处理**
   - 使用统一的响应格式: `dto.Response`
   - 返回适当的 HTTP 状态码
   - 记录详细的错误日志

6. **数据库操作**
   - 使用 GORM 进行数据库操作
   - 使用事务处理复杂操作
   - 使用软删除 (`gorm.DeletedAt`)
   - 避免 N+1 查询问题

### 前端规范

1. **项目结构**
   - `src/api/` - API 接口定义
   - `src/views/` - 页面组件
   - `src/router/` - 路由配置
   - `src/store/` - Pinia 状态管理
   - `src/locales/` - 国际化文件

2. **命名约定**
   - 文件名使用小写短横线: `user-list.vue`, `config.ts`
   - 组件名使用大驼峰: `UserList`, `ConfigForm`
   - API 方法使用小驼峰: `getUserInfo()`, `createRole()`

3. **API 调用**
   - 所有 API 调用统一在 `src/api/` 目录定义
   - 使用 TypeScript 类型定义请求和响应
   - 使用统一的错误处理

4. **国际化**
   - 所有文本必须支持中英文
   - 在 `src/locales/langs/zh-CN/` 和 `en-US/` 添加翻译
   - 使用 `$t()` 函数引用翻译键

## 开发工作流

### 启动开发环境
```bash
./dev.sh  # 一键启动前后端
```

### 添加新功能
1. 后端: Model → DTO → Validator → Service → Handler → Router → Swagger
2. 前端: API → View → Router → i18n
3. 测试功能
4. 提交代码

### API 文档
- 访问 Swagger UI: http://localhost:8080/swagger/index.html
- 导出 OpenAPI: `server/docs/swagger.json` 或 `swagger.yaml`

### 数据库迁移
- GORM 自动迁移在 `main.go` 中配置
- 种子数据在 `internal/database/seed.go`

## Git 提交规范

使用 Conventional Commits 格式:
- `feat:` - 新功能
- `fix:` - 修复 bug
- `docs:` - 文档更新
- `style:` - 代码格式调整
- `refactor:` - 重构
- `test:` - 测试相关
- `chore:` - 构建/工具链更新

示例:
```
feat: add user management API
fix: resolve JWT token expiration issue
docs: update API documentation
```

## 权限系统

- 使用 Casbin RBAC 模型
- 配置文件: `server/rbac/model.conf` 和 `policy.csv`
- 角色:
  - `super` - 超级管理员 (所有权限)
  - `admin` - 管理员 (读写系统管理)
  - `user` - 普通用户 (仅查看)

## 部署

### 开发环境
```bash
./dev.sh
```

### 生产部署
```bash
./deploy.sh  # 使用 .deploy.env 配置
```

**生产环境安全要求：**
- 必须禁用 Swagger 文档访问（在 `main.go` 中注释或条件编译 Swagger 路由）
- 使用环境变量控制 Swagger 开关，生产环境设置为 `ENABLE_SWAGGER=false`
- 确保 JWT secret 使用强随机字符串
- 配置 HTTPS 和防火墙规则

## 注意事项

1. **安全**
   - 不要在代码中硬编码敏感信息
   - 使用环境变量或配置文件
   - JWT secret 必须足够复杂
   - **生产环境必须禁用 Swagger 文档**，避免 API 信息泄露

2. **性能**
   - 使用数据库索引
   - 避免 N+1 查询
   - 合理使用缓存

3. **代码质量**
   - 保持代码简洁，避免过度工程
   - 只添加必要的功能
   - 不要添加未使用的依赖
   - 遵循 DRY 原则

4. **测试**
   - 手动测试所有新功能
   - 确保 API 文档与实现一致
   - 验证权限控制正确性

## 常用命令

### 后端
```bash
cd server
go mod tidy                    # 安装依赖
go run main.go                 # 启动服务
swag init -g main.go -o docs --parseDependency --parseInternal  # 生成 API 文档
air                            # 热重载开发
```

### 前端
```bash
cd web
pnpm install                   # 安装依赖
pnpm dev:antd                  # 启动开发服务器
pnpm build:antd                # 构建生产版本
```

## 默认账号

| 账号 | 密码 | 角色 | 权限 |
|------|------|------|------|
| vben | 123456 | super | 所有权限 |
| admin | 123456 | admin | 系统管理 |
| jack | 123456 | user | 仅查看 |

## 端口配置

- 后端 API: http://localhost:8080
- 前端开发: http://localhost:5666
- Swagger UI: http://localhost:8080/swagger/index.html (仅开发环境)
