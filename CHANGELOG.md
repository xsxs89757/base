# 更新日志

基底（base）的版本记录。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循语义化版本（规则见 README「版本与发布」）。

下游项目用 `make sync-base` 合入基底更新，然后照对应版本的「升级步骤」操作。
`make base-version` 可以查看当前已合入的基底版本和远端最新版本。

## [Unreleased]

## [2.2.0] - 2026-09-20

让 CI 少红、红了好修，并清掉 vben 上游残留。没有破坏性改动，接口和数据结构没动。

背景：下游此前没有任何本地校验（`make base-check` 在下游必然误报，且不含前端），
CI 又是拷贝继承的——基底修好的 CI 下游不同步就拿不到。khgl 因此连红 12 次
（脚手架测试在任何下游都必然失败，基底早修了它没同步），sdut 连红 7 次
（`server/docs` 过期 4 次、同一个 vue-tsc 类型错误连挂 4 次）。

### 新增

- `scripts/check.sh`：本地、pre-push 钩子、CI 跑同一份检查逻辑。
  `make check` / `check-backend` / `check-frontend` / `typecheck` 是它的转发。
  放脚本而非 make 目标：Git for Windows 不带 make，钩子调不了；Makefile 是下游的
  冲突文件，冲突解错就把目标丢了。
- `.githooks/pre-push`：push 前按改动路径挑检查（`server/` 跑后端，`admin/` 只跑
  20 秒的类型检查而不是 90 秒的构建）。`make hooks` 启用，`./dev.sh` 启动时也会
  自动启用（不覆盖已有的 husky/lefthook）。跳过：`git push --no-verify`、
  `BASE_SKIP_HOOKS=1`、`git config base.prepush off`。找不到 go/pnpm 时只警告不拦。
- `.github/workflows/checks.yml`：可复用 workflow，下游的 `ci.yml` 通过
  `@ci-v1` 调用。**修 CI 不再需要下游先同步**。按路径跳过 job（只改后端不再跑
  4 分钟前端），并带一条守卫：上次 run 不是 success 就全量跑，避免「先弄坏后端、
  再只改前端」让 run 假绿。
- 两个新挂载点，基底承诺永不创建：`.github/workflows/project.yml`（下游自有 CI job）、
  `Makefile.project`（下游自有 make 目标，带 `## 说明` 注释会进 `make help`，
  写进 `PROJECT_CHECKS` 会并入 `make check`）。
- `.gitattributes`：`server/go.sum` 用 union 合并（每次同步必撞车而两边都该保留，
  合并后自动 `go mod tidy` 清理）；脚本和钩子强制 LF，避免 Windows autocrlf 下
  钩子变成 `bash\r` 跑不起来。
- `server/config.prod.yaml.example`：此前这个文件没有模板，首次部署要等编译完、
  连上服务器才在 `cp` 那步报 "No such file or directory"。
- `admin/VBEN_PATCHES.md`：记录 `admin/` 相对上游 vben 改了什么。
- `deploy.sh` 支持 `SSH_KEY` 密钥认证；口令改用 `sshpass -e`（`-p` 会把口令暴露在
  `ps` 里）。

### 变更

- **下游不再因 `server/docs` 过期变红**：时效校验只在基底本体做。下游的 `dev.sh`
  和 `deploy.sh` 每次都会重新生成，入库那份对它们没有意义。
- `sync-base` 合并前打印「已合入版本 → 目标版本」之间每一版的升级步骤；跨大版本
  需要显式 `YES=1`；合并后 `server/go.mod` 有变化自动 `go mod tidy`。
- `deploy.sh` 在 SSH 探测和编译之前预检生产配置：文件缺失、`mode` 不是 production
  （会种下 admin/jack 的 123456 演示账号）、jwt.secret 仍是占位值都直接拒绝，
  `enable_swagger: true` 给警告。
- `AGENTS.md` 成为 AI 协作约定的唯一来源，`CLAUDE.md` 只剩一行 `@AGENTS.md`。
- swag 版本的钉死位置从四处减到三处（CI 改走 `make swagger`）。

### 移除

清理 vben 上游残留，全部零引用（共 39 个文件）：

- `admin/tea.yaml` —— tea.xyz 清单，内含 vben 作者的区块链钱包地址；
- `admin/apps/backend-mock/` —— 29 个文件、零引用、带硬编码 JWT secret，
  但在 `apps/*` 通配里，每次 CI 和每台开发机都要装它整棵依赖树；
- `admin/lefthook.yml` + `.commitlintrc.js` —— 从来没装上过的死配置
  （没有 `prepare` 脚本，`.git/hooks` 是空的）；
- `admin/README*.md` 三份上游营销文档、`.gitpod.yml`、`.gitconfig`、`.dockerignore`、
  `cspell.json`、`vben-admin.code-workspace`、`scripts/deploy/`；
- `admin/package.json` 里 15 个指向已删目录的脚本，及相应的根 devDependencies；
- `.cursor/rules/` —— 同一套规则的第三份拷贝，已漂移（还指向旧版 Vben 文档）。

`admin/packages/**` 和 `admin/internal/**` 未做任何改动（保持与上游对齐）。
穿云接入从 `dev.sh` 拆到 `scripts/dev-chuanyun.sh`，行为不变，`dev.sh` 697 行降到 472 行。

### 升级步骤

1. `.github/workflows/ci.yml` 会冲突：取基底版本（`git checkout --theirs
   .github/workflows/ci.yml`），把自己加的 job 挪到新建的
   `.github/workflows/project.yml`（模板见 AGENTS.md「六个下游挂载点」）。
   自己加的 `cmp CLAUDE.md AGENTS.md` 这类步骤要删掉。
2. `CLAUDE.md` 会冲突。**先 `diff CLAUDE.md AGENTS.md`**，把只存在于 CLAUDE.md 里的
   内容并进 AGENTS.md，然后 CLAUDE.md 取基底版本（只有一行 `@AGENTS.md`）。
3. `admin/pnpm-lock.yaml` 若冲突：取基底版本再 `cd admin && pnpm install --no-frozen-lockfile`。
4. 用到了被删内容的项目自行加回：`backend-mock`、`cspell`、`lefthook`/`commitlint`、
   `@playwright/test`、`@changesets/*`、`is-ci`，以及 `admin/package.json` 里那些
   `build:ele` / `dev:play` 之类的脚本。
5. `.cursor/rules/*` 会是 modify/delete 冲突：`git rm` 即可；想自己留着也行。
6. Makefile 里已有 `check` / `check-backend` / `check-frontend` / `check-scripts` /
   `typecheck` / `hooks` 同名目标的，改名避开；今后新目标写进 `Makefile.project`。
7. 从 dev.sh 里抠 `chuanyun_*` 函数用的测试，改成 source `scripts/dev-chuanyun.sh`。
8. 执行一次 `make hooks` 启用 pre-push 检查（跑过 `./dev.sh` 的会自动启用）。
9. 部署前确认 `server/config.prod.yaml` 的 `mode: production`、jwt.secret 不是占位值，
   否则新的预检会拒绝部署（这两条本来就会让线上出问题，只是以前没人拦）。

## [2.1.0] - 2026-09-15

钉 base-kit v1.1.0（新增 `basekit.Go`：后台任务随 app 关闭而停止），去掉 CI 上每次都有的两条注解；
接口和数据结构没动。

### 新增

- 钉 base-kit v1.1.0：`basekit.Go(app, fn)` 在 `Routes` 里起常驻后台任务（投递 worker、定时扫描），
  `app.Shutdown()` 时 ctx 取消并等任务返回；`basekit.AppContext(app)` 只取这个 ctx。
  用 `go fn(context.Background())` 起的任务永远不停，而 kit 的 `store.DB` 是包级变量：测试里一个包先后
  `NewApp` 好几次，前面 app 的任务不会退出，而是转去读写新 app 的库，同一个 worker 同时跑好几份。
  下游真实案例：7 个 e2e 测试各 `NewApp` 一次，回调投递 worker 叠了 7 份，同一条回调发两遍，
  CI 报「逐项回调 8 条，期望 4 条」，时好时坏。CLAUDE.md / AGENTS.md 补了这条约定。
  kit 顺带修了多次 `NewApp` 时 `go test -race` 偶发报的数据竞争。

### 修复

- 前端 CI 每次都有一条红色注解 `error TS4058: Return type of exported function has or is using name 'Props'`，
  job 其实是绿的，但在通知邮件和运行页上看着像前端挂了。来源是 vben 自带的 `@vben-core/tabs-ui`：
  `use-tabs-view-scroll.ts` 把 `scrollbarRef` 声明成 `InstanceType<typeof VbenScrollbar>`，构建生成 `.d.ts`
  时要引用 scrollbar.vue 里没导出的 `Props`，报错后退化成 `any`；`setup-node` 自带的 tsc 匹配规则把这行
  输出变成了注解。改成 `ComponentPublicInstance`（只用到 `$el`）。vben 上游同一处代码没修。
- CI 的 actions 升到 Node 24 版本（checkout / setup-go / setup-node v7，pnpm/action-setup v5），
  去掉每个 job 都有的「Node.js 20 is deprecated」警告。

### 升级步骤

- `make sync-base` 带来 `server/go.mod` 里的 base-kit v1.1.0。暂不同步基底、只想先拿 kit：
  `cd server && go get github.com/xsxs89757/base-kit@v1.1.0`。
- 查一遍 `Routes` 调用链里起的常驻 goroutine（`grep -rnE '^\s+go ' server/internal`），ctx 来自
  `context.Background()` 的改成 `basekit.Go`：签名是 `func(ctx context.Context)` 的直接传
  （`basekit.Go(app, worker.Run)`），其余包一层闭包。测试里 `t.Cleanup(func() { _ = app.Shutdown() })`
  要在关库的 cleanup **之后**注册（后注册的先执行），Shutdown 等任务退完再关库。
- 改过 `.github/workflows/ci.yml` 的下游，同步时这个文件大概率冲突：保留自己加的步骤，actions 版本取基底的。

## [2.0.2] - 2026-09-14

修下游 CI 必然失败、操作日志在 MySQL 上丢上传记录、`/admin` 中间件被重复挂载，钉 base-kit v1.0.4；
接口和数据结构没动。

### 修复

- CI 的「脚手架 vet & test」加 `if: github.repository == 'xsxs89757/base'`，
  与 hooks-guard 同一条件。`tools/create-base` 的 `scaffold_test.go` 里
  `fixtureBase()` 会把**当前仓库自己的文件**（含 `admin/apps/web-antd/.env`）
  拷成一份假基底再跑一遍脚手架，其中一步要把 `VITE_APP_TITLE=Admin` 替换成
  新项目名。这个耦合在基底本体是有意的——它正是用来发现「改了模板却忘了改
  脚手架」；但下游按约定必须改掉那个值（见 CLAUDE.md「新项目初始化」），
  于是替换匹配到 0 处，**这一步在任何下游仓库都是必然失败**，且与下游自己的
  代码无关。下游同步后 CI 才能真正变绿。

### 变更

- 钉 base-kit v1.0.4：
  - **操作日志在 MySQL / PostgreSQL 上记得下文件上传了**。之前上传请求的二进制请求体原样写进 `body` 列，
    utf8mb4 列（严格模式）拒绝整条 INSERT（`Error 1366`），这条日志静默丢失；SQLite 不校验编码，本地看不出来。
    超过 2KB 的中文请求（按字节截断切开汉字）、超长 path / User-Agent（`Error 1406`）同样会丢。
    上传请求的 `body` 现在是表单摘要：普通字段照录（敏感字段脱敏），文件只记文件名、大小和类型。
  - 并发时操作日志的 method / path / User-Agent 可能记成别的请求（条目引用了 Fiber 会复用的缓冲区），已修复。
  - `JWTAuth` / `PermissionAuth` / `OperationLog` 同一请求只生效一次：已经在 `/admin` 下重复挂载的下游
    不改代码也只记一条，鉴权和权限码也只判一次。

### 文档

- 下游 `/admin` 路由不要再挂 `JWTAuth` / `PermissionAuth` / `OperationLog`。kit 把这三个中间件挂在 `/admin`
  前缀上，`router/project.go` 里注册的 `/admin/*` 路由本来就经过它们；而该文件的注释让人「参考 admin.go 中
  protected 分组的中间件挂法」，照做每个 POST/PUT/DELETE 记两条操作日志，鉴权和权限码也各跑两遍。
  这条注释从挂载点引入起就是错的（当时 `Setup` 同样先调 `SetupAdmin`），不只是 2.0.1「已知」里说的路径过时。
  挂载点冻结，注释不改，以 README「权限说明」和 CLAUDE.md / AGENTS.md 为准，那里写明了哪些路由已经挂好、
  哪些要自己挂（`/api` 等其他前缀；`PreRoutes` 覆盖 kit 的接口，它排在 kit 的中间件前面，不挂就不用登录）。

### 升级步骤

- `make sync-base` 带来 `server/go.mod` 里的 base-kit v1.0.4。暂时不同步基底、只想先拿 kit 的修复：
  `cd server && go get github.com/xsxs89757/base-kit@v1.0.4`。
- 有程序按原文解析操作日志 `body` 的：上传（multipart）请求的 `body` 变成了 JSON 摘要，不再是原始请求体。
- 检查 `server/internal/router/project.go`：`/admin` 下的分组或路由上挂了 `JWTAuth()` / `PermissionAuth()` /
  `OperationLog()` 的，删掉这几个中间件（`RegisterRoutePermissions` / `RegisterAuthenticatedRoutes` 的登记保留）。
  已经写进 `sys_operation_logs` 的重复记录不会自动清理。
- 在 `main.go` 里用 `PreRoutes` 覆盖过 kit 接口的，确认覆盖的路由自己挂了 `JWTAuth()` / `PermissionAuth()`，
  写操作再加 `OperationLog()`。

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
