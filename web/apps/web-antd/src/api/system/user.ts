import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemUserApi {
  export interface SystemUser {
    [key: string]: any;
    email?: string;
    id: string;
    phone?: string;
    realName: string;
    remark?: string;
    roleIds: number[];
    roles: string[];
    status: 0 | 1;
    username: string;
  }
}

async function getUserList(params: Recordable<any>) {
  return requestClient.get<Array<SystemUserApi.SystemUser>>(
    '/system/user/list',
    { params },
  );
}

async function createUser(data: Recordable<any>) {
  return requestClient.post('/system/user', data);
}

async function updateUser(id: string, data: Recordable<any>) {
  return requestClient.put(`/system/user/${id}`, data);
}

async function deleteUser(id: string) {
  return requestClient.delete(`/system/user/${id}`);
}

export { createUser, deleteUser, getUserList, updateUser };
