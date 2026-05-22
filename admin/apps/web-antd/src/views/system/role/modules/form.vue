<script lang="ts" setup>
import type { DataNode } from 'ant-design-vue/es/tree';

import type { Recordable } from '@vben/types';

import type { SystemRoleApi } from '#/api/system/role';

import { computed, nextTick, ref } from 'vue';

import { Tree, useVbenDrawer } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { message, Spin } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { createRole, getRoleMenuTree, updateRole } from '#/api/system/role';
import { $t } from '#/locales';

import { useFormSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<SystemRoleApi.SystemRole>();

const [Form, formApi] = useVbenForm({
  schema: useFormSchema(),
  showDefaultActions: false,
});

const permissions = ref<DataNode[]>([]);
const loadingPermissions = ref(false);
// Tree 仅在菜单数据加载完成后挂载，避免 treeData 为空时
// 内部 updateTreeValue 把 modelValue 清空导致提交的 permissions 丢失。
const treeReady = ref(false);

const id = ref();
const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();

    // 防御：如果 permissions 字段意外为非数组（例如初始化异常），
    // 直接拒绝提交，避免误把所有菜单关联清空。
    if (values.permissions !== undefined && !Array.isArray(values.permissions)) {
      message.error($t('system.role.permissionLoadFailed'));
      return;
    }

    drawerApi.lock();
    (id.value ? updateRole(id.value, values) : createRole(values))
      .then(() => {
        emits('success');
        drawerApi.close();
      })
      .catch(() => {
        drawerApi.unlock();
      });
  },

  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<SystemRoleApi.SystemRole>();
      formApi.resetForm();

      if (data) {
        formData.value = data;
        id.value = data.id;
      } else {
        id.value = undefined;
      }

      // 每次打开都重新拉取菜单树，保证管理员对菜单的最新调整能反映出来。
      // 加载完成前 Tree 不会挂载，以防初始化时序问题清空已勾选项。
      treeReady.value = false;
      await loadPermissions();
      treeReady.value = true;

      // 等 Tree 用最新 treeData 完成首轮 watchEffect 后再写入 modelValue，
      // 确保 setValues 不会被 Tree 自身的"过滤无效 ID"逻辑误清空。
      await nextTick();
      await nextTick();
      if (data) {
        formApi.setValues(data);
      }
    } else {
      // 抽屉关闭时让 Tree 一并卸载，下一次打开时按最新菜单数据重建。
      treeReady.value = false;
    }
  },
});

async function loadPermissions() {
  loadingPermissions.value = true;
  try {
    // 使用角色管理专属菜单树接口，避免依赖"菜单管理"权限
    const res = await getRoleMenuTree();
    permissions.value = res as unknown as DataNode[];
  } finally {
    loadingPermissions.value = false;
  }
}

const getDrawerTitle = computed(() => {
  return formData.value?.id
    ? $t('common.edit', $t('system.role.name'))
    : $t('common.create', $t('system.role.name'));
});

function getNodeClass(node: Recordable<any>) {
  const classes: string[] = [];
  if (node.value?.type === 'button') {
    classes.push('inline-flex');
  }

  return classes.join(' ');
}
</script>
<template>
  <Drawer :title="getDrawerTitle">
    <Form>
      <template #permissions="slotProps">
        <Spin :spinning="loadingPermissions" wrapper-class-name="w-full">
          <Tree
            v-if="treeReady && permissions.length > 0"
            :tree-data="permissions"
            multiple
            bordered
            :default-expanded-level="2"
            :get-node-class="getNodeClass"
            v-bind="slotProps"
            value-field="id"
            label-field="meta.title"
            icon-field="meta.icon"
          >
            <template #node="{ value }">
              <IconifyIcon v-if="value.meta.icon" :icon="value.meta.icon" />
              {{ $t(value.meta.title) }}
            </template>
          </Tree>
        </Spin>
      </template>
    </Form>
  </Drawer>
</template>
<style lang="css" scoped>
:deep(.ant-tree-title) {
  .tree-actions {
    display: none;
    margin-left: 20px;
  }
}

:deep(.ant-tree-title:hover) {
  .tree-actions {
    display: flex;
    flex: auto;
    justify-content: flex-end;
    margin-left: 20px;
  }
}
</style>
