import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

/**
 * 注意：本项目是 backend 权限模式（main.ts 强制 accessMode: 'backend'），
 * 菜单与路由由后端 GET /admin/menu/all 提供，这里的静态路由模块不会生成菜单，
 * 也不会被注册到路由表。保留它只是为了给 IDE/类型提供参考；
 * 页面组件的可用列表来自 routes/index.ts 里对 views/ 目录的扫描（componentKeys）。
 * 新增页面：写 views/ 下的组件，再到后台"菜单管理"里配置菜单，或在
 * server/internal/store/project.go 的种子里登记。
 */
const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -1,
      title: $t('page.dashboard.title'),
    },
    name: 'Dashboard',
    path: '/dashboard',
    children: [
      {
        name: 'Workspace',
        path: '/workspace',
        component: () => import('#/views/dashboard/workspace/index.vue'),
        meta: {
          affixTab: true,
          icon: 'carbon:workspace',
          title: $t('page.dashboard.workspace'),
        },
      },
    ],
  },
];

export default routes;
