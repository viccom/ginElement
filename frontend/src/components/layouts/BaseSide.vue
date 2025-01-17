<script lang="ts" setup>
import {
  Document,
  Menu as IconMenu,
  Location,
  Setting,
} from '@element-plus/icons-vue'

// const isCollapse = ref(true)
function handleOpen(key: string, keyPath: string[]) {
  // eslint-disable-next-line no-console
  console.log(key, keyPath)
}
function handleClose(key: string, keyPath: string[]) {
  // eslint-disable-next-line no-console
  console.log(key, keyPath)
}
</script>

<template>
  <el-menu
    router
    default-active="1"
    class="el-menu-vertical-demo"
    @open="handleOpen"
    @close="handleClose"
  >
    <el-menu-item index="/nav/0">
      <el-icon>
        <Document />
      </el-icon>
      <template #title>
        应用列表
      </template>
    </el-menu-item>
    <el-menu-item index="/nav/1">
      <el-icon>
        <Document />
      </el-icon>
      <template #title>
        设备列表
      </template>
    </el-menu-item>
    <el-menu-item index="/nav/2">
      <el-icon>
        <IconMenu />
      </el-icon>
      <template #title>
        推送列表
      </template>
    </el-menu-item>
    <el-sub-menu index="">
      <template #title>
        <el-icon>
          <Location />
        </el-icon>
        <span>日志查询</span>
      </template>
      <el-menu-item-group>
        <template #title>
          <span>Group One</span>
        </template>
        <el-menu-item index="/nav/1/item-1">
          item one
        </el-menu-item>
        <el-menu-item index="1-2">
          item two
        </el-menu-item>
      </el-menu-item-group>
      <el-menu-item-group title="Group Two">
        <el-menu-item index="1-3">
          item three
        </el-menu-item>
      </el-menu-item-group>
    </el-sub-menu>
    <el-menu-item index="/nav/4">
      <el-icon>
        <Setting />
      </el-icon>
      <template #title>
        设备模板
      </template>
    </el-menu-item>
    <el-menu-item index="/nav/5">
      <el-icon>
        <Setting />
      </el-icon>
      <template #title>
        系统配置
      </template>
    </el-menu-item>
  </el-menu>
</template>
