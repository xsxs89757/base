<script lang="ts" setup>
import type {
  OnActionClickParams,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { SystemConfigApi } from '#/api/system/config';

import { onMounted, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteConfig,
  getConfigGroups,
  getConfigList,
  updateConfig,
} from '#/api/system/config';
import { $t } from '#/locales';

import { useColumns, useGridFormSchema } from './data';
import Form from './modules/form.vue';

const groupOptions = ref<Array<{ label: string; value: string }>>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: Form,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useGridFormSchema(groupOptions),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useColumns(onActionClick, onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getConfigList({
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
  } as VxeTableGridOptions<SystemConfigApi.SystemConfig>,
});

function onActionClick(e: OnActionClickParams<SystemConfigApi.SystemConfig>) {
  switch (e.code) {
    case 'delete': {
      onDelete(e.row);
      break;
    }
    case 'edit': {
      onEdit(e.row);
      break;
    }
  }
}

async function onStatusChange(
  newStatus: number,
  row: SystemConfigApi.SystemConfig,
) {
  const statusText = $t(newStatus === 1 ? 'common.enabled' : 'common.disabled');
  try {
    await new Promise((resolve, reject) => {
      Modal.confirm({
        content: $t('system.common.toggleStatusConfirm', [
          row.configKey,
          statusText,
        ]),
        onCancel: () => reject(new Error('cancelled')),
        onOk: () => resolve(true),
        title: $t('system.common.toggleStatusTitle'),
      });
    });
    await updateConfig(row.id, { ...row, status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: SystemConfigApi.SystemConfig) {
  formDrawerApi.setData(row).open();
}

function onDelete(row: SystemConfigApi.SystemConfig) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.configKey]),
    duration: 0,
    key: 'action_process_msg',
  });
  deleteConfig(row.id)
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess', [row.configKey]),
        key: 'action_process_msg',
      });
      onRefresh();
    })
    .catch(() => {
      hideLoading();
    });
}

function onRefresh() {
  gridApi.query();
  loadGroupOptions();
}

function onCreate() {
  formDrawerApi.setData({}).open();
}

async function loadGroupOptions() {
  try {
    const groups = await getConfigGroups();
    groupOptions.value = Array.isArray(groups)
      ? groups.map((g) => ({ label: g, value: g }))
      : [];
  } catch {
    groupOptions.value = [];
  }
}

onMounted(() => {
  loadGroupOptions();
});
</script>
<template>
  <Page auto-content-height>
    <FormDrawer @success="onRefresh" />
    <Grid :table-title="$t('system.config.list')">
      <template #toolbar-tools>
        <Button
          v-access:code="'System:Config:Create'"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          {{ $t('ui.actionTitle.create', [$t('system.config.name')]) }}
        </Button>
      </template>
    </Grid>
  </Page>
</template>
