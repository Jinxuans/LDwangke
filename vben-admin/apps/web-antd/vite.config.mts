import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            rewrite: (path) => path.replace(/^\/api/, '/api/v1'),
            target: 'http://localhost:8080',
            ws: true,
          },
          // PHP 通配代理：所有 .php 请求 + 旧系统公共资源
          '^.*\\.php': {
            changeOrigin: true,
            target: 'http://localhost:9000',
          },
          '/confing': {
            changeOrigin: true,
            target: 'http://localhost:9000',
          },
          '/assets': {
            changeOrigin: true,
            target: 'http://localhost:9000',
          },
        },
      },
    },
  };
});
