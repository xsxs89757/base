import type { VbenFormSchema } from '#/adapter/form';
import type { OnActionClickFn, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemUserApi } from '#/api/system/user';

import { $t } from '#/locales';

export function useFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('system.user.username'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: { type: 'password' },
      fieldName: 'password',
      label: $t('system.user.password'),
    },
    {
      component: 'Input',
      fieldName: 'realName',
      label: $t('system.user.realName'),
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: $t('system.user.email'),
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: $t('system.user.phone'),
    },
    {
      component: 'Select',
      componentProps: {
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: '请选择角色',
        style: { width: '100%' },
      },
      fieldName: 'roleIds',
      label: $t('system.user.roles'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('common.enabled'), value: 1 },
          { label: $t('common.disabled'), value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: $t('system.user.status'),
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('system.user.remark'),
    },
  ];
}

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('system.user.username'),
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
      label: $t('system.user.status'),
    },
  ];
}

export function useColumns<T = SystemUserApi.SystemUser>(
  onActionClick: OnActionClickFn<T>,
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridOptions['columns'] {
  return [
    {
      field: 'username',
      title: $t('system.user.username'),
      width: 140,
    },
    {
      field: 'realName',
      title: $t('system.user.realName'),
      width: 140,
    },
    {
      field: 'email',
      title: $t('system.user.email'),
      width: 200,
    },
    {
      field: 'phone',
      title: $t('system.user.phone'),
      width: 140,
    },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: $t('system.user.status'),
      width: 100,
    },
    {
      field: 'roles',
      formatter: ({ row }) =>
        Array.isArray(row.roles) ? row.roles.join(', ') : '',
      minWidth: 120,
      title: $t('system.user.roles'),
    },
    {
      field: 'remark',
      minWidth: 100,
      title: $t('system.user.remark'),
    },
    {
      align: 'center',
      cellRender: {
        attrs: {
          nameField: 'username',
          nameTitle: $t('system.user.name'),
          onClick: onActionClick,
        },
        name: 'CellOperation',
      },
      field: 'operation',
      fixed: 'right',
      title: $t('system.user.operation'),
      width: 130,
    },
  ];
}
