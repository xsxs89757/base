<script lang="ts" setup>
import type { SystemConfigApi } from '#/api/system/config';

import { nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createConfig, updateConfig } from '#/api/system/config';
import { $t } from '#/locales';

import { useFormSchema } from '../data';

const emits = defineEmits(['success']);

const id = ref();
const [Form, formApi] = useVbenForm({
  schema: useFormSchema(),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (id.value ? updateConfig(id.value, values) : createConfig(values))
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
      const data = drawerApi.getData<SystemConfigApi.SystemConfig>();
      formApi.resetForm();
      if (data && data.id) {
        id.value = data.id;
      } else {
        id.value = undefined;
      }
      await nextTick();
      if (data && data.id) {
        formApi.setValues(data);
      }
    }
  },
});

const getDrawerTitle = () => {
  return id.value
    ? $t('common.edit') + $t('system.config.name')
    : $t('common.create') + $t('system.config.name');
};
</script>
<template>
  <Drawer :title="getDrawerTitle()">
    <Form />
  </Drawer>
</template>
