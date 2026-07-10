import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  const apiPort = process.env.VITE_API_PORT || '8080';
  // dev.sh 自动换端口时通过 VITE_ADMIN_PORT 指定前端端口；
  // 未设置时沿用 .env.development 的 VITE_PORT (vben 只从 .env 文件读取)
  const adminPort = Number(process.env.VITE_ADMIN_PORT) || 0;
  return {
    application: {
      nitroMock: false,
    },
    vite: {
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
