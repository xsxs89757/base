<script lang="ts" setup>
import type { MenuRecordRaw } from '@vben/types';

import { computed } from 'vue';
import { useRoute } from 'vue-router';

import { Page, VbenAvatar } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';

import { $t } from '#/locales';

defineOptions({ name: 'Workspace' });

const userStore = useUserStore();
const accessStore = useAccessStore();
const route = useRoute();

function greetingPeriod(hour: number) {
  if (hour < 6 || hour >= 18) return 'evening';
  if (hour < 12) return 'morning';
  return 'afternoon';
}

const greeting = computed(() => {
  const period = greetingPeriod(new Date().getHours());
  return $t(`workspace.greeting.${period}`, [
    userStore.userInfo?.realName || userStore.userInfo?.username || '',
  ]);
});

const roleText = computed(
  () => (userStore.userInfo?.roles ?? []).join(', ') || '-',
);

/** 除工作台本身外是否还有可访问的菜单：一个都没有时提示找管理员分配权限 */
function hasOtherMenu(menus: MenuRecordRaw[]): boolean {
  return menus.some((menu) =>
    menu.children?.length
      ? hasOtherMenu(menu.children)
      : !!menu.path && menu.path !== route.path,
  );
}

const hasMenus = computed(() => hasOtherMenu(accessStore.accessMenus));
</script>

<template>
  <Page content-class="flex flex-col gap-5">
    <div class="card-box flex items-center gap-4 p-5">
      <VbenAvatar
        :src="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
        class="size-16"
      />
      <div class="flex flex-col gap-1">
        <h1 class="text-xl font-semibold">{{ greeting }}</h1>
        <span class="text-foreground/80">{{ $t('workspace.welcome') }}</span>
        <span class="text-sm text-foreground/60">
          {{ $t('workspace.currentRole') }}: {{ roleText }} ·
          {{ userStore.userInfo?.username }}
        </span>
      </div>
    </div>

    <div v-if="!hasMenus" class="card-box p-5 text-center text-foreground/60">
      {{ $t('workspace.noMenu') }}
    </div>
  </Page>
</template>
