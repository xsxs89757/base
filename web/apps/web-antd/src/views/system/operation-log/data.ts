import type { VbenFormSchema } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemOperationLogApi } from '#/api/system/operation-log';

import { $t } from '#/locales';

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('system.operationLog.username'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: 'POST', value: 'POST' },
          { label: 'PUT', value: 'PUT' },
          { label: 'DELETE', value: 'DELETE' },
        ],
      },
      fieldName: 'method',
      label: $t('system.operationLog.method'),
    },
    {
      component: 'Input',
      fieldName: 'path',
      label: $t('system.operationLog.path'),
    },
  ];
}

export function useColumns(
  onActionClick: OnActionClickFn<SystemOperationLogApi.OperationLog>,
): VxeTableGridOptions<SystemOperationLogApi.OperationLog>['columns'] {
  return [
    {
      field: 'username',
      title: $t('system.operationLog.username'),
      width: 120,
    },
    {
      cellRender: {
        name: 'CellTag',
        options: [
          { color: 'warning', label: 'POST', value: 'POST' },
          { color: 'success', label: 'PUT', value: 'PUT' },
          { color: 'error', label: 'DELETE', value: 'DELETE' },
        ],
      },
      field: 'method',
      title: $t('system.operationLog.method'),
      width: 100,
    },
    {
      field: 'path',
      minWidth: 200,
      title: $t('system.operationLog.path'),
    },
    {
      cellRender: {
        name: 'CellTag',
        options: [
          { color: 'success', label: '200', value: 200 },
          { color: 'warning', label: '400', value: 400 },
          { color: 'error', label: '401', value: 401 },
          { color: 'error', label: '403', value: 403 },
          { color: 'error', label: '500', value: 500 },
        ],
      },
      field: 'status',
      title: $t('system.operationLog.status'),
      width: 90,
    },
    {
      field: 'duration',
      formatter: ({ row }) => `${row.duration}ms`,
      title: $t('system.operationLog.duration'),
      width: 100,
    },
    {
      field: 'ip',
      title: $t('system.operationLog.ip'),
      width: 140,
    },
    {
      field: 'createTime',
      title: $t('system.operationLog.createTime'),
      width: 180,
    },
    {
      align: 'center',
      cellRender: {
        attrs: {
          nameField: 'path',
          nameTitle: $t('system.operationLog.name'),
          onClick: onActionClick,
        },
        name: 'CellOperation',
        options: ['delete'],
      },
      field: 'operation',
      fixed: 'right',
      title: $t('system.operationLog.operation'),
      width: 100,
    },
  ];
}
