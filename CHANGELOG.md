# 更新日志

基底（base）的版本记录。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循语义化版本（规则见 README「版本与发布」）。

下游项目用 `make sync-base` 合入基底更新，然后照对应版本的「升级步骤」操作。
`make base-version` 可以查看当前已合入的基底版本和远端最新版本。

## [Unreleased]

## [2.0.1] - 2026-09-02

对 v2.0.0 的评审修复，全是工程细节，接口和数据结构没动。

### 修复

- `deploy.sh` 生成 Swagger 时加 `GOWORK=off`，与下面的 `go build` 一致。本地 `make kit-dev`
  留下 `server/go.work` 时，swag 会按本地 kit 源码生成注解，而二进制按 go.mod 钉死的版本编译，
  发布出去的 docs 与实际接口对不上。
- `make base-check` 全程 `GOWORK=off`（vet / test / 交叉编译 / Swagger 时效），同样的原因：
  有 go.work 时它检查的不是将要发布的那份代码。`make swagger` / `build` / `test` 不变——
  kit-dev 时开发者就是想按本地 kit 编。
- `make migrate-kit` 去掉 `@latest`，改用 go.mod 里钉的 kit 版本。`@latest` 在 GOPROXY 有延迟的
  几分钟里会拿到旧版，而 base-kit v1.0.1 正好带着「改写 `base/internal/store`」那个 bug，
  跑完 `main.go` 直接 `undefined: store.ProjectModels`。命令失败时补了一句提示。
- `server/main_test.go` 结束时关闭 sqlite 连接：Windows 上句柄不放开，`t.TempDir()` 的清理会失败
  （dev.sh 支持 Git Bash）。
- `server/go.sum` 清掉 `go mod tidy` 会删的陈旧记录，下游第一次 tidy 不会再带上无关 diff。

### 变更

- 钉 base-kit v1.0.3：**生产模式下 5xx 响应不再回显内部错误原文**，对外统一返回
  `Internal Server Error`，原文进日志。recover 中间件会把 panic 转成同类错误走这条路径，
  之前 DB 错误、文件路径、panic 文本都能被外部看到。开发模式不变，4xx 不受影响。

### 文档

- 明确 **Go ≥ 1.25**。v2.0.0 起由 base-kit 的依赖决定（`x/sys`、`x/text`、`gorm.io/driver/postgres`
  等都声明 `go 1.25.0`），当时没写进 README。
- `go get -u github.com/xsxs89757/base-kit` 一律改成 `go get github.com/xsxs89757/base-kit@latest`：
  `-u` 会把 kit 的依赖一并升到最新 minor，与模板钉死版本的初衷相反。
- 下游业务代码的目录用 `internal/<层>/<业务名>/`，不要用 `admin`（kit 的包名）。
  放进与 kit 同名的目录，将来 kit 再接管点什么就会出现「同一 import 路径两个来源」。

### 已知

- 两个冻结挂载点里的注释还指向已搬到 kit 的 `admin.go` / `store.go`。挂载点受 `make check-hooks`
  冻结，这次不动；下次确需改动挂载点时一并修（届时会更新 `FROZEN_BLOBS`，下游有一处冲突）。

### 升级步骤

`make sync-base` 即可，没有额外动作。已经迁到 2.0.0 的项目不用再跑 `make migrate-kit`。

## [2.0.0] - 2026-09-02

后端框架层搬到独立仓库 [base-kit](https://github.com/xsxs89757/base-kit)，从此框架层的 bug 修复和新功能
走 `go get ...@latest`，不再靠 git merge 一个文件一个文件地合。`server/` 里只剩装配代码、两个挂载点和业务代码。

### 破坏性变更

- **导入路径全变**：`base/config` 和 `base/internal/{dto,validator,model,middleware,service/admin,handler/admin}`
  搬到 `github.com/xsxs89757/base-kit/*`（去掉 `internal` 前缀）。用 `make migrate-kit` 自动改写。
  `base/internal/store` 和 `base/internal/router` **不变**：这两个包在本仓库仍然存在
  （数据层与路由两个挂载点），`store.DB`、`store.IsUniqueViolation` 的写法一个字都不用改。
- **`store.Init` 换签名**：`Init(store.Options) error`，不再 `log.Fatal`；模型和种子数据由参数传入。
  下游一般不直接调用它（`basekit.Run` 负责）。
- **`main.go` 重写**：只剩 `basekit.Run(basekit.Options{...})` 和 Swagger 挂载。
- **swag 命令换参数**：`--parseDependencyLevel 3 --packagePrefix base,github.com/xsxs89757/base-kit`
  （`--parseDependency` 只解析模型不解析路由，kit 的接口会全部丢失）。Makefile / dev.sh / deploy.sh / CI 已同步。
- 生成的 `swagger.json` 路径与定义**与 v1.0.1 完全一致**，前端不受影响。

### 新增

- `basekit.Options` 扩展点：`Models` / `Seed` / `Routes`（业务路由）/ `PreRoutes`（覆盖 kit 接口）/
  `Swagger` / `Fiber` / `Config`（测试注入）。
- `make migrate-kit` 改写导入路径；`make kit-dev` / `make kit-undev` 切换「用本地 ../base-kit 源码开发」。
- `server/main_test.go` 冒烟测试：临时 sqlite 启动 → 登录 → 用户信息 → 菜单 → Swagger。
- `deploy.sh` 用 `GOWORK=off` 编译，本地的 `go.work` 不会污染发布产物。

### 升级前先看一眼

如果你往 `internal/handler/admin/`、`internal/model/admin/`、`internal/middleware/` 这类
基底目录里加过自己的文件，同步时基底只删自己那份，你的文件会留在原地。改写把引用方指向了 kit，
你那些函数就会 `undefined`，报错信息看不出根因。`make migrate-kit` 结束时会把这些目录列出来。
处理办法是把它们挪到自己的包（如 `internal/handler/biz/`）再改引用方的 import。
真实项目里见过一个下游在 6 个基底目录中放了 48 个自己的文件，先看一眼能省很多时间。

另外：如果你的项目当初是靠拷文件建的（没有和基底的共同 git 历史），`git merge` 会报
「拒绝合并无关的历史」，`make sync-base` 用不了，只能照着本节手工移植。

### 已知约束

- 给 `sys_users` 等基底表加列时，扩展结构**只能声明表名、主键和新列**，不能嵌入 `adminmodel.User`：
  嵌入会把 `Roles` many2many 带过来，往共享的 `user_roles` 表加一列 `<结构名>_id`，
  通过嵌入结构写入时 `user_id` 为 NULL，kit 的 handler 按 `user_id` 查角色会静默失效。
  详见 base-kit 的 README 与 `store/embed_test.go`。

### 升级步骤

1. `make sync-base`。预期冲突：
   - `server/go.mod` / `go.sum`：保留 `module base` 一行，两边的 require 都留下，稍后 `go mod tidy` 收拾；
   - `server/main.go`：如果改过，取基底版本再把自己的定制搬到 `basekit.Options` 的回调里；
   - `server/docs/*`：取基底版本，第 3 步会重新生成；
   - 改过 kit 已接管的核心文件时会出现 modify/delete 冲突：删掉本地版本，把改动改成给 kit 提 PR，
     或用 `basekit.Options.PreRoutes` 在本项目里覆盖。
2. `make migrate-kit`（改写导入路径并 `go mod tidy`）。
3. `make swagger && make test && make build`。
4. `make dev` 用 super 登录走一遍用户/角色/菜单页，再打一个自己的业务接口确认权限码仍然生效。

整条路径已经在一个带业务代码（模型 + 接口 + 路由权限码 + 种子菜单）的下游项目上跑通：
merge 无冲突、改写 4 个文件、编译测试通过、Swagger 30 条路径不变、
启动后下游菜单正常种下、业务接口 super 通过 / 无权限角色 403。

## [1.0.1] - 2026-09-02

### 修复

- 脚手架：`create-base <项目名> --skip-install` 这种「项目名在前、选项在后」的写法之前会被
  当成两个项目名而报错（Go 的 flag 包在第一个位置参数处停止解析），现在两种顺序都可以。

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

[Unreleased]: https://github.com/xsxs89757/base/compare/v2.0.1...HEAD
[2.0.1]: https://github.com/xsxs89757/base/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/xsxs89757/base/compare/v1.0.1...v2.0.0
[1.0.1]: https://github.com/xsxs89757/base/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/xsxs89757/base/releases/tag/v1.0.0
