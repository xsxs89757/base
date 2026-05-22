import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemRoleApi {
  export interface SystemRole {
    [key: string]: any;
    code: string;
    id: string;
    menuIds?: number[];
    name: string;
    // permissions 可选：仅在编辑权限抽屉提交时携带，
    // 状态切换等场景必须省略以避免覆盖后端已有的菜单关联。
    permissions?: number[];
    remark?: string;
    status: 0 | 1;
  }
}

/**
 * 获取角色列表数据
 */
async function getRoleList(params: Recordable<any>) {
  return requestClient.get<Array<SystemRoleApi.SystemRole>>(
    '/system/role/list',
    { params },
  );
}

/**
 * 创建角色
 * @param data 角色数据
 */
async function createRole(data: Omit<SystemRoleApi.SystemRole, 'id'>) {
  return requestClient.post('/system/role', data);
}

/**
 * 更新角色
 *
 * @param id 角色 ID
 * @param data 角色数据
 */
async function updateRole(
  id: string,
  data: Omit<SystemRoleApi.SystemRole, 'id'>,
) {
  return requestClient.put(`/system/role/${id}`, data);
}

/**
 * 删除角色
 * @param id 角色 ID
 */
async function deleteRole(id: string) {
  return requestClient.delete(`/system/role/${id}`);
}

async function getAllRoles() {
  return requestClient.get<Array<SystemRoleApi.SystemRole>>('/system/role/all');
}

/**
 * 获取角色授权用的菜单树
 *
 * 与 /system/menu/list 解耦，只要拥有 System:Role:List 权限即可调用，
 * 避免管理员被取消"菜单管理"权限后无法继续维护角色授权。
 */
async function getRoleMenuTree() {
  return requestClient.get<Array<Record<string, any>>>(
    '/system/role/menu-tree',
  );
}

export {
  createRole,
  deleteRole,
  getAllRoles,
  getRoleList,
  getRoleMenuTree,
  updateRole,
};
