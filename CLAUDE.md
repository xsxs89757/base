# AGENTS.md

本文件给 Codex/Agents 使用，作用范围为 `96_server/` 整个目录。开发时同时参考 `CLAUDE.md`，但以本文件中的路径和执行约定为当前仓库准则。

## 项目概览

这是一个后台管理系统：

- 后端：`server/`，Go Fiber v2 + GORM + Casbin v3 + JWT + Swagger/OpenAPI。
- 前端：`web/`，Vben Admin v5 + Vue 3 + TypeScript + Ant Design Vue + Vite + Pinia + Tailwind CSS。
- 前端应用主路径：`web/apps/web-antd/src/`。
- 后端管理接口前缀：`/admin`。
- 前端开发代理：`/api` 会被 Vite 改写到后端 `/admin`，见 `web/apps/web-antd/vite.config.mts`。

## 工作原则

- 先读现有实现，再改代码。优先复用项目已有的 handler/service/model、Vben adapter、页面结构和命名风格。
- 涉及完整功能时，后端、前端、路由、权限、i18n、Swagger 文档一起处理，不只改单层。
- 保持改动范围小。不要顺手重构无关模块，不要引入未使用依赖。
- 不要提交或依赖本地敏感配置：`.deploy.env`、`server/config.yaml`、`server/config.prod.yaml`。
- `server/docs/` 是 Swagger 生成物；只有 API 变更需要同步生成时才更新。

## 后端开发规则

### 分层路径

后台管理功能按下面顺序补齐：

1. `server/internal/model/admin/`：GORM 数据模型。
2. `server/internal/dto/admin/`：请求/响应 DTO。
3. `server/internal/validator/admin/`：请求校验。
4. `server/internal/service/admin/`：业务逻辑。
5. `server/internal/handler/admin/`：HTTP handler 和 Swagger 注解。
6. `server/internal/router/admin.go`：注册 `/admin/*` 路由。
7. `server/docs/`：API 变更后重新生成 Swagger。

公共层：

- `server/internal/dto/base.go` 提供统一响应：`dto.Success`、`dto.PageSuccess`、`dto.Fail`。
- `server/internal/middleware/` 放 JWT、Casbin、操作日志等中间件。
- `server/internal/store/` 放数据库和共享存储初始化。

### API 和响应

- Handler 返回统一响应格式，不直接拼零散 JSON。
- 列表接口使用分页结构：`items` + `total`。
- 管理端业务接口放在 `/admin` 前缀下；公共前台 API 才放 `/api`。
- 新增、修改、删除管理端接口必须确认 JWT、Casbin 和操作日志是否应该覆盖。
- id=1 的用户是超级管理员：不受普通权限限制，不出现在普通用户列表，不允许被修改或删除。

### Swagger

每个 handler 方法都要有完整 Swagger 注解，至少包括：

- `@Summary`
- `@Tags`
- `@Accept`
- `@Produce`
- `@Param`
- `@Success`
- `@Failure`
- `@Router`

Tags 命名保持业务可读：

- 后台：`认证`、`用户`、`系统管理 - 角色` 等。
- 前台：使用 `前台 - xxx`。

API 变更后在 `server/` 下执行：

```bash
swag init -g main.go -o docs --parseDependency --parseInternal
```

### GORM 注意事项

- 文件名使用小写下划线，例如 `operation_log.go`。
- 包名使用小写单词，例如 `admin`、`handler`。
- Go 结构体和方法使用大驼峰。
- 含连续大写缩写的字段必须显式指定列名，例如：
  - `OID` 需要 `gorm:"column:oid"`。
  - `AIPDFPath` 需要 `gorm:"column:ai_pdf_path"`。
- `map[string]any` 做 `Updates` 时，key 必须是数据库列名，也就是 snake_case，不是 Go 字段名。
- 改 model 字段后需要完整重启后端，让 AutoMigrate 重新执行；只看前端热更新不够。
- 复杂写操作使用事务，查询注意预加载和索引，避免 N+1。

## 前端开发规则

### Vben 文档查阅

- 本项目是 Vben Admin v5。遇到 table、form、drawer、modal、menu、permission、route 等通用后台组件，不要凭记忆手写，先查 Vben 文档。
- 用户指定的入口是 `https://doc.vvbin.cn/`。如果该站点提示当前页是旧版本并给出 V5 文档入口，继续查 V5 文档 `https://doc.vben.pro/`。
- 常用 V5 组件关键词：`useVbenVxeGrid`、`useVbenForm`、`useVbenDrawer`、`Vben Modal`、`Vben Vxe Table`。
- 查完文档后还要对照本项目 adapter：`#/adapter/vxe-table`、`#/adapter/form`，以仓库内封装为最终落地方式。

### 前端路径

主要目录在 `web/apps/web-antd/src/`：

- `api/`：接口定义，统一使用 `requestClient`。
- `views/`：页面组件。
- `router/`：路由配置。
- `store/`：Pinia 状态。
- `adapter/`：Vben 表单、表格、组件适配器。
- `locales/langs/zh-CN/` 和 `locales/langs/en-US/`：国际化。

系统管理页面优先沿用现有模块结构：

- `list.vue`：列表页，使用 `Page` + `useVbenVxeGrid`。
- `data.ts`：列配置、搜索表单 schema、表单 schema。
- `modules/form.vue`：新增/编辑抽屉，使用 `useVbenDrawer` + `useVbenForm`。

### 组件使用

- 列表和表格优先使用 `useVbenVxeGrid`，从 `#/adapter/vxe-table` 导入。
- 表单优先使用 `useVbenForm`，从 `#/adapter/form` 导入。
- 新增/编辑侧滑层优先使用 `useVbenDrawer`，从 `@vben/common-ui` 导入。
- 页面容器优先使用 `Page`。
- 图标优先使用项目已有图标包，例如 `@vben/icons`。
- 不要直接手写一套 Ant Design Vue Table/Form/Drawer，除非 Vben 封装明确无法满足，并在代码中保持局部化。

### 表格和表单约定

- 表格请求放在 `gridOptions.proxyConfig.ajax.query` 中，分页参数映射为后端需要的 `page`、`pageSize`。
- 表格行主键设置 `rowConfig.keyField`。
- 工具栏使用 `toolbarConfig`，自定义按钮放到 `#toolbar-tools` 或文档约定插槽。
- 搜索表单配置放在 `formOptions.schema`，常用 schema 写在同模块 `data.ts`。
- `destroyOnClose: true` 的 Drawer 中，新增和编辑打开时都必须调用 `formApi.setValues()` 或等效初始化逻辑，否则 Select、Switch 等组件容易残留旧值。
- 下拉选项优先通过 `formApi.updateSchema()` 更新 `options`，不要随意绕过项目现有 adapter。
- Switch 映射后端整数时设置 `checkedValue: 1`、`unCheckedValue: 0`。

### API、路由和 i18n

- API 方法写在 `web/apps/web-antd/src/api/`，按业务模块分文件并补 TypeScript namespace/types。
- `requestClient` 已经配置 `codeField: code`、`dataField: data`、`successCode: 0`，后端响应要与其匹配。
- 前端调用路径写逻辑路径，例如 `/system/user/list`；开发时 Vite 会把 `/api` 代理到后端 `/admin`。
- 所有页面可见文案都要走 `$t()`，并同步维护 `zh-CN` 和 `en-US` 翻译。
- 新页面需要同步路由、菜单和权限编码；动态菜单数据由后端 `/admin/menu/all` 提供。

## 验证命令

按改动范围选择验证：

```bash
# 后端
cd server
go test ./...
swag init -g main.go -o docs --parseDependency --parseInternal

# 前端
cd web
pnpm dev:antd
pnpm build:antd
```

如果只是文档或规则变更，至少检查 Markdown 内容和路径是否与当前仓库一致。

## 常用命令

```bash
# 一键启动前后端
./dev.sh

# 后端开发
cd server
go run main.go
air

# 前端开发
cd web
pnpm install
pnpm dev:antd
pnpm build:antd
```

## 默认账号

| 账号 | 密码 | 角色 | 权限 |
| --- | --- | --- | --- |
| super | 123456 | super | 所有权限 |
| admin | 123456 | admin | 系统管理 |
| jack | 123456 | user | 仅查看 |

## 端口

- 后端 API：`http://localhost:8080`
- 前端开发：`http://localhost:5666`
- Swagger UI：`http://localhost:8080/swagger/index.html`，仅开发环境使用
