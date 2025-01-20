import path from 'node:path'

import Vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import VueRouter from 'unplugin-vue-router/vite'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },

  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "~/styles/element/index.scss" as *;`,
        api: 'modern-compiler',
      },
    },
  },

  plugins: [
    Vue({
      include: [/\.vue$/, /\.md$/],
    }),
    // https://github.com/posva/unplugin-vue-router
    VueRouter({
      extensions: ['.vue'],
      dts: 'src/typed-router.d.ts',
    }),

    Components({
      // allow auto load markdown components under `./src/components/`
      extensions: ['vue', 'md'],
      // allow auto import and register components used in markdown
      include: [/\.vue$/, /\.vue\?vue/, /\.md$/],
      resolvers: [
        ElementPlusResolver({
          importStyle: 'sass',
        }),
      ],
      dts: 'src/components.d.ts',
    }),

    // https://github.com/antfu/unocss
    // see uno.config.ts for config
    Unocss(),
  ],

  ssr: {
    // TODO: workaround until they support native ESM
    noExternal: ['element-plus'],
  },

  server: {
    proxy: {
      // 代理所有以 /api 开头的请求
      '/api': {
        target: 'http://127.0.0.1:8880', // 后端服务地址
        changeOrigin: true, // 允许跨域
        rewrite: path => path.replace(/^\/api/, '/api'), // 保留 /api
      },
      // 代理 Swagger UI 的所有资源
      '/swagger': {
        target: 'http://localhost:8880', // 后端服务地址
        changeOrigin: true, // 允许跨域
      },
    },

    port: 3330, // 开发服务器端口
    open: true, // 自动打开浏览器
  },
  build: {
    outDir: '../backend/html', // 编译输出目录
    assetsDir: 'assets', // 静态资源目录
  },
})
