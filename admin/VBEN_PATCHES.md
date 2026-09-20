# admin/ 与上游 Vben Admin 的差异

`admin/` 是 [vue-vben-admin](https://github.com/vbenjs/vue-vben-admin) 的 in-tree 副本
（1000+ 文件，占整个仓库的 96%）。本文件记录它与上游的差异，以及将来怎么升级。

**基底与下游的约定：`admin/packages/**` 和 `admin/internal/**` 尽量不要改。** 不是因为改不动，
而是改了就得在这里登记，否则下一次升级 vben 时没人知道哪些改动必须保住。业务代码请写在
`admin/apps/web-antd/src/`。

## 基线

**上游 v5.6.0（2026-02-09，commit `8a215fbc`）**，`admin/package.json` 的 version 也是 5.6.0。

核对方式（`packages/` 与 `internal/`，忽略 `dist`、`.turbo`、`node_modules`）：

```bash
git clone --depth 1 --branch v5.6.0 https://github.com/vbenjs/vue-vben-admin.git /tmp/vben
diff -rq /tmp/vben/packages admin/packages
diff -rq /tmp/vben/internal admin/internal
```

当前共 **44 处差异**：40 个文件内容不同 + 4 个本仓库独有路径。分三类，性质完全不同。

## 一、本仓库自己的改动（必须保住，共 6 个文件）

只有这些是 base 仓库提交历史里真正改过的（`git log -- admin/packages admin/internal`）：

| 文件 | 提交 | 改了什么 |
| --- | --- | --- |
| `packages/@core/ui-kit/tabs-ui/src/use-tabs-view-scroll.ts` | `d4cf2bd` | `scrollbarRef` 改用 `ComponentPublicInstance`（只用到 `$el`），否则 `vue-tsc` 构建报 TS4058。上游同一处没修。 |
| `packages/locales/src/langs/zh-CN/authentication.json`<br>`packages/locales/src/langs/en-US/authentication.json` | `da533b2` | 登录页标题简化，去掉 Vben 版权页脚。 |
| `packages/stores/src/modules/access.ts`<br>`packages/stores/src/modules/access.test.ts` | `3310aa5` `0c88f84` | 菜单权限码 RBAC：菜单可见性与按钮权限的判定，对齐后端 `/admin/menu/all` 的数据形状。 |
| `packages/@core/preferences/src/config.ts`<br>`packages/@core/preferences/__tests__/__snapshots__/config.test.ts.snap` | `bc93481` | 默认偏好项调整（配合权限改造）。 |

升级时这 6 个文件的改动必须逐条重放。

## 二、导入时从更新版上游回移的（4 个路径）

这些在 v5.6.0 里没有、在上游后续版本里有，是 2026-03-25 导入时一并带进来的：

| 路径 | 上游何时有的 |
| --- | --- |
| `packages/@core/ui-kit/layout-ui/src/hooks/use-sidebar-drag.ts` | `afffc4b3`，2026-02-27，侧边栏拖拽 |
| `packages/@core/ui-kit/tabs-ui/src/components/widgets/tool-refresh.vue` | v5.6.0 之后 |
| `packages/effects/layouts/src/hooks/` | v5.6.0 之后 |
| `packages/effects/layouts/src/route-cached/` | v5.6.0 之后 |

升级到更高版本的上游时，这些会自然被上游版本覆盖，**不需要**手工保留。

## 三、其余 ~34 个文件内容差异

同样来自导入那一刻，不是本仓库改的：tabs-ui、menu-ui、layout-ui、preferences 抽屉、
locales 的 common/preferences、`internal/vite-config/src/index.ts` 等。它们是上游
v5.6.0 到 2026-03 之间自身的演进（导入的是 main 的一个快照，只是 `package.json`
仍写着 5.6.0）。

判断依据：本仓库的提交历史从没碰过这些文件，而上游 main 上能找到对应内容
（例如 `packages/@core/base/icons/src/lucide.ts` 的 blob 与上游 `2930dcd7` 一致）。

**升级时这一类直接取上游版本即可。**

## 升级 vben 的步骤

```bash
# 1. 取上游两个版本
git clone https://github.com/vbenjs/vue-vben-admin.git /tmp/vben && cd /tmp/vben

# 2. 整目录替换（基线 -> 目标版本），本仓库独有的东西先不管
git archive <目标tag> packages internal | tar -x -C <base仓库>/admin

# 3. 重放第一节那 6 个文件的改动（逐条看上面的表）

# 4. 验证
cd admin
pnpm install --no-frozen-lockfile     # 上游依赖变了就要重新生成锁文件
pnpm -F @vben/web-antd typecheck
pnpm build:antd
```

替换后务必重新核对一遍第一节：`access.ts` 的权限逻辑和 `use-tabs-view-scroll.ts`
的类型修复最容易被覆盖掉，前者会让菜单权限失效（页面能进但不该进），后者会让构建报错。

跨大版本（如 5.x → 6.x）时上游常有 break change，别整目录替换，照上游 CHANGELOG 走。

## 已知可以从上游回移的能力

- **富文本编辑器 Tiptap**：上游 `bb78882f`（2026-03-30）加的
  `packages/effects/plugins/src/tiptap/`，比本仓库导入时间晚 5 天，所以基底没有。
  下游 96agent 和 laitui 各自回移了一份（各约 2085 行，位置相同），khgl 也在做富文本。
  这是共性需求，应当由基底统一回移一次，登记到本文件，而不是每个项目各搞一套。
