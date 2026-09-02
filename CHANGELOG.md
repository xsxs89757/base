# 更新日志

基底（base）的版本记录。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循语义化版本（规则见 README「版本与发布」）。

下游项目用 `make sync-base` 合入基底更新，然后照对应版本的「升级步骤」操作。
`make base-version` 可以查看当前已合入的基底版本和远端最新版本。

## [Unreleased]

## [1.0.0] - 2026-09-02

第一个正式版本。此前下游合的是 `base/main` 的 HEAD，没有版本概念；从本版本起基底按语义化版本发布。

### 破坏性变更

- **token 增加类型标记（typ）**：access 与 refresh token 不再互通。升级后所有已登录用户需要重新登录一次。
- **移除 Casbin**：此前 `CasbinAuth` 从未真正调用 `Enforcer.Enforce`，权限判定一直走「路由 → 权限码 → 角色菜单」。
  `middleware.CasbinAuth()` 保留为 `PermissionAuth()` 的别名，下游代码无需改动；旧库里的 `casbin_rule` 表不再使用。
- **角色 / 用户 / 配置改为物理删除**：它们的 name/code/username/config_key 带唯一索引，软删行会永久占住键值，
  导致「删过的角色再也建不出同 code 的」。启动时自动清理这三张表里历史遗留的软删除行。菜单、部门仍是软删除。
- **`mode: production` 真正生效**：`jwt.secret` 必须至少 32 位且不能是示例占位值，否则拒绝启动；
  生产库只创建 `super` 账号（不再建 admin / jack 演示账号）。

### 新增

- **版本化发布**：`vX.Y.Z` 标签、仓库根 `.base-version`、`make base-release` / `make base-version`。
- **`make sync-base` 默认合入最新基底标签**，`make sync-base VERSION=v1.2.0` 或 `VERSION=main` 可指定。
  基底标签在下游以 `base/v*` 命名空间存在，不与下游自己的标签冲突。
- **脚手架**：`go run github.com/xsxs89757/base/tools/create-base@latest <项目名>` 一条命令创建新项目
  （clone 保留共同历史、生成随机 jwt.secret、按项目名填好 `.deploy.env` 与前端 `.env`、装依赖、提交首个 commit）。
- **GitHub Actions CI**：后端 vet/test/交叉编译、前端 typecheck/build、Swagger 文档一致性检查；
  挂载点冻结校验（`make check-hooks`）只在基底仓库本体运行。
- JWTAuth 每个请求以数据库为准核对用户状态与启用角色（进程内缓存 1 分钟），禁用用户、调整角色、
  修改密码立即生效；改密后此前签发的 token 全部作废。
- 非 super 操作者不得修改或删除持有 super 角色的用户。
- 菜单 / 部门递归删除并校验父级不能是自身或下级；列表分页统一上限 200；唯一冲突返回 400、记录不存在返回 404。
- 角色列表支持按创建时间区间筛选。
- 前端：`/workspace` 欢迎页按当前用户可访问菜单生成快捷入口，替换 vben 演示页（analytics / profile / about）。
- `dev.sh` 挂 EXIT trap，异常退出时也关闭隧道与子进程。

### 修复

- 删除前端 `index.html` 里 vben 模板自带的百度统计脚本（会把后台每个页面 URL 上报到第三方账号）。
- 菜单表单「激活路径」绑定错误导致编辑即丢失，现在可填可清空。
- 用户物理删除时先删 `user_roles` 再删主行，避免 MySQL / PostgreSQL 上的外键报错。
- 演示菜单 Analytics / About 启动时按 name + component 精确移除，不会误删下游同名菜单。
- `server/internal/router/project.go` 注释里过时的 Casbin 说明。**这是挂载点最后一次改动**，
  此后由 `make check-hooks` 冻结；下游若改写过该文件顶部注释，本次同步会有一处注释冲突，取任意一方即可。

### 升级步骤

1. **先设置 tag 拉取策略**（重要）：老版本 `make sync-base` 里的 `git fetch base` 会把基底标签直接拉成
   下游的裸标签 `v1.0.0`，与下游自己的版本号冲突：

   ```bash
   git config remote.base.tagOpt --no-tags
   git tag -d v1.0.0 2>/dev/null   # 如果已经被拉进来了
   ```

2. `make sync-base`（这一步用的还是老 Makefile，等价于合并 `base/main`；合入后即为 v1.0.0，
   之后再同步就是新语义：默认合最新标签）。
3. 旧库执行 `DROP TABLE casbin_rule;`（可选，只是清理）。
4. 通知所有用户重新登录（token 格式变了）。
5. 检查生产 `config.yaml`：`mode: production` 时 `jwt.secret` 必须 ≥32 位且不是示例占位值，
   否则服务拒绝启动。生成方式：`openssl rand -base64 48`。

[Unreleased]: https://github.com/xsxs89757/base/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/xsxs89757/base/releases/tag/v1.0.0
