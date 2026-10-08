<script lang="ts" setup>
import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { SystemOperationLogApi } from '#/api/system/operation-log';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';
import { createIconifyIcon } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  clearOperationLog,
  deleteOperationLog,
  getOperationLogList,
} from '#/api/system/operation-log';
import { $t } from '#/locales';

import { useColumns, useGridFormSchema } from './data';

const Trash2 = createIconifyIcon('lucide:trash-2');

// 审计记录只有超管能删（后端同样只认超管），普通管理员即便有删除权限码也不显示删除/清空
const { hasAccessByRoles } = useAccess();
const canDelete = hasAccessByRoles(['super']);

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useGridFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useColumns(onActionClick, canDelete),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getOperationLogList({
            page: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      search: true,
      zoom: true,
    },
  } as VxeTableGridOptions<SystemOperationLogApi.OperationLog>,
});

function onActionClick(
  e: OnActionClickParams<SystemOperationLogApi.OperationLog>,
) {
  if (e.code === 'delete') {
    onDelete(e.row);
  }
}

function onDelete(row: SystemOperationLogApi.OperationLog) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.path]),
    duration: 0,
    key: 'action_process_msg',
  });
  deleteOperationLog(row.id)
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess', [row.path]),
        key: 'action_process_msg',
      });
      gridApi.query();
    })
    .catch(() => {
      hideLoading();
    });
}

function onClear() {
  Modal.confirm({
    content: $t('system.operationLog.clearConfirm'),
    onOk: async () => {
      await clearOperationLog();
      message.success($t('system.operationLog.clearSuccess'));
      gridApi.query();
    },
    title: $t('system.operationLog.clearTitle'),
  });
}
</script>
<template>
  <Page auto-content-height>
    <Grid :table-title="$t('system.operationLog.list')">
      <template #toolbar-tools>
        <Button v-if="canDelete" danger type="primary" @click="onClear">
          <Trash2 class="size-4" />
          {{ $t('system.operationLog.clear') }}
        </Button>
      </template>
    </Grid>
  </Page>
</template>
