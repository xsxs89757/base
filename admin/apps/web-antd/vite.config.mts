import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  const apiPort = process.env.VITE_API_PORT || '8080';
  // dev.sh 自动换端口时通过 VITE_ADMIN_PORT 指定前端端口；
  // 未设置时沿用 .env.development 的 VITE_PORT (vben 只从 .env 文件读取)
  const adminPort = Number(process.env.VITE_ADMIN_PORT) || 0;

  // 穿云内网穿透：给 dev server 一个固定公网地址（微信回调、手机真机、演示给人看）。
  // 三层降级，任何一层出问题都只是没有公网地址，本地开发照常：
  //   1) 包没装（老项目同步基底后未重新 pnpm install）-> 动态 import 失败，跳过
  //   2) CHUANYUN=0 -> 插件自身跳过
  //   3) 穿云客户端没开/没登录 -> 插件内部静默降级
  // 隧道名由 dev.sh 按项目注入 CHUANYUN_NAME，直接跑 pnpm dev:antd 时回落到 package.json 名称。
  const plugins = [];
  if (process.env.CHUANYUN !== '0') {
    try {
      const { default: chuanyun } = await import('vite-plugin-chuanyun');
      plugins.push(chuanyun({ name: process.env.CHUANYUN_NAME || undefined }));
    } catch {
      // 未安装 vite-plugin-chuanyun，跳过
    }
  }

  return {
    application: {
      nitroMock: false,
    },
    vite: {
      plugins,
      server: {
        ...(adminPort ? { port: adminPort, strictPort: true } : {}),
        proxy: {
          '/api': {
            changeOrigin: true,
            rewrite: (path: string) => path.replace(/^\/api/, '/admin'),
            target: `http://localhost:${apiPort}`,
            ws: true,
          },
        },
      },
    },
  };
});
