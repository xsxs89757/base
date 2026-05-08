import type { VbenFormSchema } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemConfigApi } from '#/api/system/config';

import { $t } from '#/locales';

export function useFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'configName',
      label: $t('system.config.configName'),
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'configKey',
      label: $t('system.config.configKey'),
      rules: 'required',
    },
    {
      component: 'Switch',
      componentProps: {
        checkedChildren: $t('common.enabled'),
        checkedValue: 1,
        unCheckedChildren: $t('common.disabled'),
        unCheckedValue: 0,
      },
      defaultValue: 1,
      fieldName: 'status',
      label: $t('system.config.status'),
    },
    {
      component: 'Textarea',
      componentProps: {
        rows: 6,
      },
      fieldName: 'configValue',
      label: $t('system.config.configValue'),
      rules: 'required',
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('system.config.remark'),
    },
  ];
}

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'configKey',
      label: $t('system.config.configKey'),
    },
    {
      component: 'Input',
      fieldName: 'configGroup',
      label: $t('system.config.configGroup'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('common.enabled'), value: 1 },
          { label: $t('common.disabled'), value: 0 },
        ],
      },
      fieldName: 'status',
      label: $t('system.config.status'),
    },
  ];
}

export function useColumns<T = SystemConfigApi.SystemConfig>(
  onActionClick: OnActionClickFn<T>,
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions['columns'] {
  return [
    {
      field: 'configName',
      title: $t('system.config.configName'),
      width: 160,
    },
    {
      field: 'configKey',
      title: $t('system.config.configKey'),
      width: 200,
    },
    {
      field: 'configValue',
      minWidth: 200,
      title: $t('system.config.configValue'),
    },
    {
      field: 'configGroup',
      title: $t('system.config.configGroup'),
      width: 120,
    },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: $t('system.config.status'),
      width: 100,
    },
    {
      field: 'remark',
      title: $t('system.config.remark'),
      width: 200,
    },
    {
      field: 'createTime',
      title: $t('system.config.createTime'),
      width: 180,
    },
    {
      align: 'center',
      cellRender: {
        attrs: {
          nameField: 'configKey',
          nameTitle: $t('system.config.name'),
          onClick: onActionClick,
        },
        name: 'CellOperation',
      },
      field: 'operation',
      fixed: 'right',
      title: $t('system.config.operation'),
      width: 130,
    },
  ];
}
