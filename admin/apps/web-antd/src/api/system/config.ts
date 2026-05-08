import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemConfigApi {
  export interface SystemConfig {
    [key: string]: any;
    configGroup: string;
    configKey: string;
    configName: string;
    configValue: string;
    createTime?: string;
    id: string;
    remark?: string;
    status: 0 | 1;
  }
}

async function getConfigList(params: Recordable<any>) {
  return requestClient.get<Array<SystemConfigApi.SystemConfig>>(
    '/system/config/list',
    { params },
  );
}

async function createConfig(data: Omit<SystemConfigApi.SystemConfig, 'id'>) {
  return requestClient.post('/system/config', data);
}

async function updateConfig(
  id: string,
  data: Omit<SystemConfigApi.SystemConfig, 'id'>,
) {
  return requestClient.put(`/system/config/${id}`, data);
}

async function deleteConfig(id: string) {
  return requestClient.delete(`/system/config/${id}`);
}

async function getConfigGroups(): Promise<string[]> {
  return requestClient.get('/system/config/groups');
}

export { createConfig, deleteConfig, getConfigGroups, getConfigList, updateConfig };
