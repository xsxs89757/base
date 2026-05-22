import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it } from 'vitest';

import { useAccessStore } from './access';

describe('useAccessStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('updates accessMenus state', () => {
    const store = useAccessStore();
    expect(store.accessMenus).toEqual([]);
    store.setAccessMenus([{ name: 'Dashboard', path: '/dashboard' }]);
    expect(store.accessMenus).toEqual([
      { name: 'Dashboard', path: '/dashboard' },
    ]);
  });

  it('updates accessToken state correctly', () => {
    const store = useAccessStore();
    expect(store.accessToken).toBeNull(); // 初始状态
    store.setAccessToken('abc123');
    expect(store.accessToken).toBe('abc123');
  });

  it('returns the correct accessToken', () => {
    const store = useAccessStore();
    store.setAccessToken('xyz789');
    expect(store.accessToken).toBe('xyz789');
  });

  // 测试设置空的访问菜单列表
  it('handles empty accessMenus correctly', () => {
    const store = useAccessStore();
    store.setAccessMenus([]);
    expect(store.accessMenus).toEqual([]);
  });

  // 测试设置空的访问路由列表
  it('handles empty accessRoutes correctly', () => {
    const store = useAccessStore();
    store.setAccessRoutes([]);
    expect(store.accessRoutes).toEqual([]);
  });

  it('resets generated access state without clearing the token', () => {
    const store = useAccessStore();
    store.setAccessToken('abc123');
    store.setAccessCodes(['System:Role:List']);
    store.setAccessMenus([{ name: 'System', path: '/system' }]);
    store.setAccessRoutes([{ name: 'System', path: '/system' }]);
    store.setIsAccessChecked(true);

    store.resetAccessState();

    expect(store.accessToken).toBe('abc123');
    expect(store.accessCodes).toEqual([]);
    expect(store.accessMenus).toEqual([]);
    expect(store.accessRoutes).toEqual([]);
    expect(store.isAccessChecked).toBe(false);
  });
});
