import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
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
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
