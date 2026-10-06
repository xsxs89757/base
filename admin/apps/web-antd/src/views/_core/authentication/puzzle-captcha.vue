<script lang="ts" setup>
import type { AuthApi } from '#/api';

import { computed, onMounted, reactive, ref, useTemplateRef } from 'vue';

import { SliderCaptcha } from '@vben/common-ui';
import { RotateCw } from '@vben/icons';
import { $t } from '@vben/locales';

import { useElementSize } from '@vueuse/core';

import { getCaptchaApi, verifyCaptchaApi } from '#/api';

/**
 * 拼图滑块验证码，外观照 vben 的 SliderTranslateCaptcha，拖动条直接用它的 SliderCaptcha。
 * 区别在于缺口位置由后端生成、由后端判定：vben 原组件在浏览器里随机缺口、自己判断通过，
 * 绕过页面直接调登录接口就失效了。
 */
const props = defineProps<{ initial?: AuthApi.CaptchaResult }>();
const emit = defineEmits<{ success: [token: string] }>();

const boxRef = useTemplateRef('boxRef');
const barRef = useTemplateRef('barRef');
const { width: boxWidth } = useElementSize(boxRef);

const puzzle = ref<AuthApi.CaptchaResult | undefined>(props.initial);
// 滑块条的 v-model：is-slot 模式下为 false 时松手会自动弹回，校验期间要让它停在原处
const passed = ref(false);
const state = reactive({
  dragging: false,
  failText: '',
  moveX: 0,
  result: '' as '' | 'fail' | 'success',
  startTime: 0,
  time: '',
  verifying: false,
});

// 显示宽度 / 原图宽度：拼图块按同一比例缩放，提交时把位移换算回原图像素
const scale = computed(() =>
  puzzle.value?.width ? boxWidth.value / puzzle.value.width : 1,
);

const boxStyle = computed(() => {
  const { height, width } = puzzle.value ?? {};
  return { aspectRatio: width && height ? `${width} / ${height}` : '2 / 1' };
});

const pieceStyle = computed(() => {
  const { height, pieceSize, pieceY = 0, width } = puzzle.value ?? {};
  if (!width || !height || !pieceSize) return { display: 'none' };
  return {
    left: `${state.moveX}px`,
    top: `${(pieceY / height) * 100}%`,
    width: `${(pieceSize / width) * 100}%`,
  };
});

const successText = computed(() =>
  state.verifying
    ? $t('page.auth.captchaVerifying')
    : $t('ui.captcha.sliderSuccessText'),
);

async function refresh() {
  if (state.verifying) return;
  state.result = '';
  state.moveX = 0;
  passed.value = false;
  barRef.value?.resume();
  try {
    puzzle.value = await getCaptchaApi();
  } catch {
    // 失败提示由 requestClient 的响应拦截器统一弹出
  }
}

function handleStart() {
  state.result = '';
  state.startTime = Date.now();
}

function handleMove({ moveX }: { moveX: number }) {
  state.dragging = true;
  state.moveX = Math.max(0, moveX);
}

async function handleEnd() {
  const moved = state.dragging && state.moveX > 2;
  state.dragging = false;
  const captchaId = puzzle.value?.captchaId;
  // 只是点了一下没拖：不提交（拼图只能提交一次，误点不该把它作废）
  if (!moved || !captchaId || state.verifying) {
    state.moveX = 0;
    return;
  }
  passed.value = true;
  state.verifying = true;
  try {
    const resp = await verifyCaptchaApi({
      captchaId,
      x: state.moveX / scale.value,
    });
    const token = resp.data.data?.token;
    if (!token) throw new Error('missing token');
    state.time = ((Date.now() - state.startTime) / 1000).toFixed(1);
    state.result = 'success';
    emit('success', token);
  } catch (error: any) {
    state.failText =
      error?.response?.data?.message || $t('ui.captcha.sliderTranslateFailTip');
    state.result = 'fail';
    state.verifying = false;
    state.moveX = 0;
    passed.value = false;
    barRef.value?.resume();
    // 拼图提交一次即作废：停一下让人看到失败提示，再换一张
    setTimeout(refresh, 800);
    return;
  }
  state.verifying = false;
}

onMounted(() => {
  if (!props.initial?.image) refresh();
});

defineExpose({ refresh });
</script>

<template>
  <div class="flex w-full select-none flex-col items-center">
    <div
      ref="boxRef"
      class="relative w-full overflow-hidden rounded-md border border-border shadow-md"
      :style="boxStyle"
    >
      <template v-if="puzzle?.image">
        <img
          :src="puzzle.image"
          alt=""
          class="block size-full cursor-pointer"
          draggable="false"
          @click="refresh"
        />
        <img
          :src="puzzle.piece"
          alt=""
          class="pointer-events-none absolute"
          :style="pieceStyle"
          draggable="false"
        />
      </template>
      <button
        :aria-label="$t('ui.captcha.refreshAriaLabel')"
        :title="$t('ui.captcha.refreshAriaLabel')"
        class="absolute right-2 top-2 rounded-full bg-black/30 p-1 text-white hover:bg-black/50"
        type="button"
        @click="refresh"
      >
        <RotateCw class="size-4" />
      </button>
      <div
        class="absolute bottom-0 left-0 w-full text-center text-xs leading-[30px] text-white"
      >
        <div v-if="state.result === 'success'" class="bg-success/80">
          {{ $t('ui.captcha.sliderTranslateSuccessTip', [state.time]) }}
        </div>
        <div v-else-if="state.result === 'fail'" class="bg-destructive/80">
          {{ state.failText }}
        </div>
        <div v-else-if="!state.dragging" class="bg-black/30">
          {{ $t('ui.captcha.sliderTranslateDefaultTip') }}
        </div>
      </div>
    </div>
    <SliderCaptcha
      ref="barRef"
      v-model="passed"
      class="mt-4"
      is-slot
      :success-text="successText"
      @end="handleEnd"
      @move="handleMove"
      @start="handleStart"
    />
  </div>
</template>
