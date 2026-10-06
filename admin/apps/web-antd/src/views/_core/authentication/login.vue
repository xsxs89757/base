<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import type { AuthApi } from '#/api';

import { computed, ref } from 'vue';

import { AuthenticationLogin, useVbenModal, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { getCaptchaApi } from '#/api';
import { useAuthStore } from '#/store';

import PuzzleCaptcha from './puzzle-captcha.vue';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();

// 后台「系统配置」login_captcha 开启时，点登录先弹拼图，拖对了再带着一次性凭证去登录
const [CaptchaModal, captchaModalApi] = useVbenModal({
  centered: true,
  class: 'w-[380px] max-w-[calc(100vw-32px)]',
  footer: false,
  fullscreenButton: false,
});
const pendingPuzzle = ref<AuthApi.CaptchaResult>();
let pendingValues: Recordable<any> | undefined;

async function login(params: Recordable<any>) {
  try {
    await authStore.authLogin(params);
  } catch {
    // 失败提示由 requestClient 的响应拦截器统一弹出
  }
}

async function handleLogin(values: Recordable<any>) {
  let captcha: AuthApi.CaptchaResult;
  try {
    captcha = await getCaptchaApi();
  } catch {
    return;
  }
  if (!captcha.enabled) {
    await login(values);
    return;
  }
  pendingValues = values;
  pendingPuzzle.value = captcha;
  captchaModalApi.open();
}

async function handleCaptchaSuccess(captchaToken: string) {
  // 留一下「验证成功」的提示再关
  await new Promise((resolve) => setTimeout(resolve, 500));
  captchaModalApi.close();
  const values = pendingValues;
  pendingValues = undefined;
  if (values) await login({ ...values, captchaToken });
}

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.password'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
  ];
});
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    :show-code-login="false"
    :show-forget-password="false"
    :show-qrcode-login="false"
    :show-register="false"
    :show-third-party-login="false"
    @submit="handleLogin"
  />
  <CaptchaModal :title="$t('ui.captcha.title')">
    <PuzzleCaptcha :initial="pendingPuzzle" @success="handleCaptchaSuccess" />
  </CaptchaModal>
</template>
