<script lang="ts" setup>
import type { Tab } from '~/utils/tabUtils'
import { ElMessageBox } from 'element-plus'
import { onMounted, ref, watch } from 'vue' // 引入 watch
import { addTab, handleTabsEdit, initDefaultTab, restoreTabsFromLocalStorage, saveTabsToLocalStorage } from '~/utils/tabUtils'
import ChartComponent from './1_components/ChartComponent.vue'
import TableDevList from './1_components/TableDevList.vue'
import TabledevView from './1_components/TabledevView.vue'

// 定义标签页类型
type TabType = 'devs' | 'tags' | 'devView' | 'chart'

// 指定唯一的 storageKey
const devPage_storageKey = 'devtabs'

// 动态组件映射
const componentMap = {
  devs: TableDevList,
  devView: TabledevView,
  chart: ChartComponent,
}

// 定义 activeTab 和 tabs
const activeTab = ref('1')
const tabs = ref<Tab[]>(initDefaultTab({
  title: '总览',
  name: '1',
  type: 'devs',
  config: {
    apiUrl: '/api/v1/listDevices',
  },
  jsonData: {
    devName: '',
    devDesc: '',
    devId: '',
    devType: '',
    instId: '',
  },
}))

// 在页面加载时恢复TAB标签页状态
onMounted(() => {
  const { tabs: savedTabs, activeTab: savedActiveTab } = restoreTabsFromLocalStorage(devPage_storageKey)
  if (savedTabs.length > 0) {
    tabs.value = savedTabs
    activeTab.value = savedActiveTab
  }
})
// 监听 activeTab 的变化
watch(activeTab, (newActiveTab) => {
  // 保存TAB状态
  saveTabsToLocalStorage(tabs.value, newActiveTab, devPage_storageKey)
})

// 添加新标签页
function addNewTab(title: string, type: TabType, config: { apiUrl: string }, jsonData: { [key: string]: any }) {
  const result = addTab(tabs.value, activeTab.value, title, type, config, jsonData)
  tabs.value = result.tabs
  activeTab.value = result.activeTab
  saveTabsToLocalStorage(tabs.value, activeTab.value, devPage_storageKey)
}

// 处理“数据”按钮点击事件
function handleDataClick(devName: string, devDesc: string, devId: string, instId: string) {
  const tabName = `查看[${devName}]`
  addNewTab(tabName, 'devView', {
    apiUrl: `/api/v1/getDevvalues`,
  }, {
    devName,
    devDesc,
    devId,
    instId,
  })
}

// 处理“点表”按钮点击事件
function handlePointClick(devName: string, devDesc: string, devId: string, instId: string) {
  const tabName = `点表[${devName}]`
  addNewTab(tabName, 'tags', { // 修改类型为'tags'
    apiUrl: '/api/v1/getDevtags', // 新增API路径
  }, {
    devName,
    devDesc,
    devId,
    instId,
  })
}

// 处理“删除”按钮点击事件
function handleDeleteClick(devName: string, devId: string) {
  ElMessageBox.alert(`删除设备：<br>设备名称=${devName}<br>设备ID=${devId}`, '提示', {
    confirmButtonText: '确定',
    dangerouslyUseHTMLString: true, // 允许使用 HTML 字符串
  })
}

// 处理标签页关闭
function handleCloseTab(targetName: string) {
  const result = handleTabsEdit(tabs.value, activeTab.value, targetName, devPage_storageKey)
  tabs.value = result.tabs
  activeTab.value = result.activeTab
  // 保存TAB状态
  saveTabsToLocalStorage(tabs.value, activeTab.value, devPage_storageKey)
}
</script>

<template>
  <div>
    设备列表
  </div>
  <el-tabs
    v-model="activeTab"
    type="card"
    editable
    class="demo-tabs"
    @tab-remove="handleCloseTab"
  >
    <el-tab-pane
      v-for="(tab, index) in tabs"
      :key="tab.name"
      :label="tab.title"
      :name="tab.name"
      :closable="index !== 0"
    >
      <!-- 动态加载组件 -->
      <component
        :is="componentMap[tab.type]"
        :config="tab.config"
        :json-data="tab.jsonData"
        @data-click="handleDataClick"
        @point-click="handlePointClick"
        @delete-click="handleDeleteClick"
      />
    </el-tab-pane>
  </el-tabs>
</template>

<style>
.demo-tabs > .el-tabs__content {
  padding: 32px;
  color: #6b778c;
  font-size: 32px;
  font-weight: 600;
}
</style>
