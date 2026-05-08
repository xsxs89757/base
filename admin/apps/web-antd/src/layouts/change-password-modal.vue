<script lang="ts" setup>
import { reactive, ref } from 'vue';

import { Form, Input, message, Modal } from 'ant-design-vue';

import { requestClient } from '#/api/request';

const open = ref(false);
const loading = ref(false);

const formState = reactive({
  confirmPassword: '',
  newPassword: '',
  oldPassword: '',
});

const formRef = ref();

function show() {
  formState.oldPassword = '';
  formState.newPassword = '';
  formState.confirmPassword = '';
  open.value = true;
}

async function handleOk() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  loading.value = true;
  try {
    await requestClient.post('/auth/change-password', {
      newPassword: formState.newPassword,
      oldPassword: formState.oldPassword,
    });
    message.success('密码修改成功');
    open.value = false;
  } catch (error: any) {
    message.error(error?.response?.data?.message || '密码修改失败');
  } finally {
    loading.value = false;
  }
}

function validateConfirm(_rule: any, value: string) {
  if (value && value !== formState.newPassword) {
    return Promise.reject(new Error('两次输入的密码不一致'));
  }
  return Promise.resolve();
}

defineExpose({ show });
</script>

<template>
  <Modal
    v-model:open="open"
    :confirm-loading="loading"
    destroy-on-close
    title="修改密码"
    @ok="handleOk"
  >
    <Form
      ref="formRef"
      :label-col="{ span: 5 }"
      :model="formState"
      :wrapper-col="{ span: 18 }"
      class="pt-4"
    >
      <Form.Item
        label="旧密码"
        name="oldPassword"
        :rules="[{ message: '请输入旧密码', required: true }]"
      >
        <Input.Password
          v-model:value="formState.oldPassword"
          placeholder="请输入旧密码"
        />
      </Form.Item>
      <Form.Item
        label="新密码"
        name="newPassword"
        :rules="[
          { message: '请输入新密码', required: true },
          { message: '密码长度不能少于6位', min: 6 },
        ]"
      >
        <Input.Password
          v-model:value="formState.newPassword"
          placeholder="请输入新密码（至少6位）"
        />
      </Form.Item>
      <Form.Item
        label="确认密码"
        name="confirmPassword"
        :rules="[
          { message: '请再次输入新密码', required: true },
          { validator: validateConfirm },
        ]"
      >
        <Input.Password
          v-model:value="formState.confirmPassword"
          placeholder="请再次输入新密码"
        />
      </Form.Item>
    </Form>
  </Modal>
</template>
