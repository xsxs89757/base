import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemOperationLogApi {
  export interface OperationLog {
    [key: string]: any;
    createTime?: string;
    duration: number;
    id: string;
    ip: string;
    method: string;
    path: string;
    status: number;
    userAgent?: string;
    username: string;
  }
}

async function getOperationLogList(params: Recordable<any>) {
  return requestClient.get<Array<SystemOperationLogApi.OperationLog>>(
    '/system/operation-log/list',
    { params },
  );
}

async function deleteOperationLog(id: string) {
  return requestClient.delete(`/system/operation-log/${id}`);
}

async function clearOperationLog() {
  return requestClient.delete('/system/operation-log/clear');
}

export { clearOperationLog, deleteOperationLog, getOperationLogList };
