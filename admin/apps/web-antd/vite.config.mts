import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  const apiPort = process.env.VITE_API_PORT || '8080';
  return {
    application: {
      nitroMock: false,
    },
    vite: {
      server: {
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
