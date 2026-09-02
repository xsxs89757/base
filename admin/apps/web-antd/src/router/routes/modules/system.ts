import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

/**
 * 注意：backend 权限模式下菜单与路由来自后端 GET /admin/menu/all，本文件不会生成菜单，
 * 改这里加不出新菜单。新增系统页面：写 views/system 下的组件，再到"菜单管理"配置，
 * 或在 server/internal/store/store.go 的 seedMenus 里登记。详见 dashboard.ts 头部说明。
 */
const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'ion:settings-outline',
      order: 9997,
      title: $t('system.title'),
    },
    name: 'System',
    path: '/system',
    children: [
      {
        path: '/system/user',
        name: 'SystemUser',
        meta: {
          icon: 'mdi:account-outline',
          title: $t('system.user.title'),
        },
        component: () => import('#/views/system/user/list.vue'),
      },
      {
        path: '/system/role',
        name: 'SystemRole',
        meta: {
          icon: 'mdi:account-group',
          title: $t('system.role.title'),
        },
        component: () => import('#/views/system/role/list.vue'),
      },
      {
        path: '/system/menu',
        name: 'SystemMenu',
        meta: {
          icon: 'mdi:menu',
          title: $t('system.menu.title'),
        },
        component: () => import('#/views/system/menu/list.vue'),
      },
      {
        path: '/system/dept',
        name: 'SystemDept',
        meta: {
          icon: 'charm:organisation',
          title: $t('system.dept.title'),
        },
        component: () => import('#/views/system/dept/list.vue'),
      },
      {
        path: '/system/config',
        name: 'SystemConfig',
        meta: {
          icon: 'carbon:settings-adjust',
          title: $t('system.config.title'),
        },
        component: () => import('#/views/system/config/list.vue'),
      },
      {
        path: '/system/operation-log',
        name: 'SystemOperationLog',
        meta: {
          icon: 'carbon:activity',
          title: $t('system.operationLog.title'),
        },
        component: () => import('#/views/system/operation-log/list.vue'),
      },
    ],
  },
];

export default routes;
