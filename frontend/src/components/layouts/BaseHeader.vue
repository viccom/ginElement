<script lang="ts" setup>
import { ref } from 'vue'
import { repository } from '~/../package.json'
import { toggleDark } from '~/composables'

// 定义响应式数据
const apiDocUrl = ref('/swagger/index.html')

// 定义方法
function openAboutPage() {
  window.open(apiDocUrl.value, '_blank')
}
</script>

<template>
  <el-menu class="el-menu-demo" mode="horizontal" :ellipsis="false" router>
    <el-menu-item index="/">
      <div class="flex items-center justify-center gap-2">
        <div class="text-xl" i-ep-element-plus />
        <span>GoIOT</span>
      </div>
    </el-menu-item>

    <el-menu-item>
      <a class="size-full flex items-center justify-center no-underline" @click="openAboutPage">
        API
      </a>
    </el-menu-item>

    <el-menu-item index="/help/1">
      Help
    </el-menu-item>

    <el-menu-item h="full" @click="toggleDark()">
      <button
        class="w-full cursor-pointer border-none bg-transparent"
      >
        <i inline-flex i="dark:ep-moon ep-sunny" />
      </button>
    </el-menu-item>

    <el-menu-item h="full">
      <a class="size-full flex items-center justify-center" :href="repository.url" target="_blank">
        <div i-ri-github-fill />
      </a>
    </el-menu-item>
  </el-menu>
</template>

<style lang="scss">
.el-menu-demo {
  &.ep-menu--horizontal > .ep-menu-item:nth-child(1) {
    margin-right: auto;
  }

  // 添加以下规则以去掉超链接下划线
  .no-underline {
    text-decoration: none;
  }
}
</style>
