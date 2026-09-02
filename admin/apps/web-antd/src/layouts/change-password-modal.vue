<script lang="ts" setup>
import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { changePasswordApi } from '#/api';
import { $t } from '#/locales';
import { useAuthStore } from '#/store';

const authStore = useAuthStore();

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
  },
  schema: [
    {
      component: 'InputPassword',
      defaultValue: '',
      fieldName: 'oldPassword',
      label: $t('system.password.oldPassword'),
      rules: 'required',
    },
    {
      component: 'InputPassword',
      defaultValue: '',
      fieldName: 'newPassword',
      label: $t('system.password.newPassword'),
      rules: z
        .string()
        .min(
          6,
          $t('ui.formRules.minLength', [$t('system.password.newPassword'), 6]),
        ),
    },
    {
      component: 'InputPassword',
      defaultValue: '',
      dependencies: {
        rules(values) {
          return z
            .string()
            .min(
              1,
              $t('ui.formRules.required', [
                $t('system.password.confirmPassword'),
              ]),
            )
            .refine(
              (value) => value === values.newPassword,
              $t('authentication.confirmPasswordTip'),
            );
        },
        triggerFields: ['newPassword'],
      },
      fieldName: 'confirmPassword',
      label: $t('system.password.confirmPassword'),
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const { newPassword, oldPassword } = await formApi.getValues();
    modalApi.lock();
    try {
      // 失败提示由 requestClient 的响应拦截器统一弹出，这里不重复 toast
      await changePasswordApi({ newPassword, oldPassword });
      message.success($t('system.password.changeSuccess'));
      modalApi.close();
      // 后端改密后已让旧 token 失效，直接回登录页重新登录
      await authStore.logout();
    } finally {
      modalApi.unlock();
    }
  },
});
</script>
<template>
  <Modal :title="$t('common.changePassword')">
    <Form class="mx-4" />
  </Modal>
</template>
