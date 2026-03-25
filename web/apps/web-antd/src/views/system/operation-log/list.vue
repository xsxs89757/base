<script lang="ts" setup>
import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { SystemOperationLogApi } from '#/api/system/operation-log';

import { Page } from '@vben/common-ui';
import { createIconifyIcon } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

const Trash2 = createIconifyIcon('lucide:trash-2');

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  clearOperationLog,
  deleteOperationLog,
  getOperationLogList,
} from '#/api/system/operation-log';
import { $t } from '#/locales';

import { useColumns, useGridFormSchema } from './data';

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useGridFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useColumns(onActionClick),
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
    content: '确定要清空所有操作日志吗？此操作不可恢复。',
    onOk: async () => {
      await clearOperationLog();
      message.success('操作日志已清空');
      gridApi.query();
    },
    title: '清空操作日志',
  });
}
</script>
<template>
  <Page auto-content-height>
    <Grid :table-title="$t('system.operationLog.list')">
      <template #toolbar-tools>
        <Button danger type="primary" @click="onClear">
          <Trash2 class="size-4" />
          {{ $t('system.operationLog.clear') }}
        </Button>
      </template>
    </Grid>
  </Page>
</template>
