<script lang="ts" setup>
import type { WorkbenchQuickNavItem } from '@vben/common-ui';
import type { MenuRecordRaw } from '@vben/types';

import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Page, VbenAvatar, WorkbenchQuickNav } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import { $t } from '#/locales';

defineOptions({ name: 'Workspace' });

const userStore = useUserStore();
const accessStore = useAccessStore();
const router = useRouter();
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

/** 递归收集可访问的叶子菜单（排除当前页），作为快捷入口 */
function collectLeaves(menus: MenuRecordRaw[], out: MenuRecordRaw[] = []) {
  for (const menu of menus) {
    if (menu.children?.length) {
      collectLeaves(menu.children, out);
    } else if (menu.path && menu.path !== route.path) {
      out.push(menu);
    }
  }
  return out;
}

const quickNavItems = computed<WorkbenchQuickNavItem[]>(() =>
  collectLeaves(accessStore.accessMenus).map((menu) => ({
    icon: menu.icon || 'carbon:application',
    title: $t(menu.name),
    url: menu.path,
  })),
);

function navTo(item: WorkbenchQuickNavItem) {
  if (!item.url) return;
  if (item.url.startsWith('http')) {
    openWindow(item.url);
    return;
  }
  router.push(item.url);
}
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

    <WorkbenchQuickNav
      v-if="quickNavItems.length > 0"
      :items="quickNavItems"
      :title="$t('workspace.quickNav')"
      @click="navTo"
    />
    <div v-else class="card-box p-5 text-center text-foreground/60">
      {{ $t('workspace.noMenu') }}
    </div>
  </Page>
</template>
