<script lang="ts" setup>
import type { SystemUserApi } from '#/api/system/user';

import { nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { getAllRoles } from '#/api/system/role';
import { createUser, updateUser } from '#/api/system/user';
import { $t } from '#/locales';

import { useFormSchema } from '../data';

const emits = defineEmits(['success']);

const id = ref();
const roleOptions = ref<Array<{ label: string; value: number }>>([]);

const [Form, formApi] = useVbenForm({
  schema: useFormSchema(),
  showDefaultActions: false,
});

async function loadRoleOptions() {
  try {
    const roles = await getAllRoles();
    roleOptions.value = (roles || []).map((r) => ({
      label: r.name,
      value: Number(r.id),
    }));
  } catch {
    roleOptions.value = [];
  }
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (id.value ? updateUser(id.value, values) : createUser(values))
      .then(() => {
        emits('success');
        drawerApi.close();
      })
      .catch(() => {
        drawerApi.unlock();
      });
  },
  async onOpenChange(isOpen) {
    if (!isOpen) return;

    const data = drawerApi.getData<SystemUserApi.SystemUser>();
    const isEdit = !!(data && data.id);
    id.value = isEdit ? data.id : undefined;

    if (roleOptions.value.length === 0) {
      await loadRoleOptions();
    }

    formApi.updateSchema([
      {
        componentProps: { disabled: isEdit },
        fieldName: 'username',
      },
      {
        componentProps: {
          placeholder: isEdit ? '留空则不修改' : undefined,
          type: 'password',
        },
        fieldName: 'password',
        rules: isEdit ? undefined : 'required',
      },
      {
        componentProps: {
          options: roleOptions.value,
        },
        fieldName: 'roleIds',
      },
    ]);

    await nextTick();
    if (isEdit) {
      formApi.setValues(data);
    } else {
      formApi.setValues({});
    }
  },
});

const getDrawerTitle = () => {
  return id.value
    ? $t('common.edit') + $t('system.user.name')
    : $t('common.create') + $t('system.user.name');
};
</script>
<template>
  <Drawer :title="getDrawerTitle()">
    <Form />
  </Drawer>
</template>
