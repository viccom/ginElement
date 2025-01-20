<script lang="ts" setup>
import type { Tab } from '~/utils/tabUtils'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, ref, watch } from 'vue' // 引入 watch
import { addTab, handleTabsEdit, initDefaultTab, restoreTabsFromLocalStorage, saveTabsToLocalStorage } from '~/utils/tabUtils'
import AppEditor from './0_components/AppEditor.vue'
import AppList from './0_components/AppList.vue'

// 动态组件映射
const componentMap = {
  appList: AppList,
  appEditor: AppEditor,
}

// 定义 activeTab 和 tabs
const activeTab = ref('1')
const tabs = ref<Tab[]>(initDefaultTab({
  title: '总览',
  name: '1',
  type: 'appList',
  config: {
    apiUrl: `/api/v1/listApps?t=${Date.now()}`,
  },
  jsonData: {
    instName: '',
    instId: '',
    isRunning: false,
  },
}))

// 指定唯一的 storageKey
const appPage_storageKey = 'apptabs'

// 在页面加载时恢复TAB标签页状态
onMounted(() => {
  const { tabs: savedTabs, activeTab: savedActiveTab } = restoreTabsFromLocalStorage(appPage_storageKey)
  if (savedTabs.length > 0) {
    tabs.value = savedTabs
    activeTab.value = savedActiveTab
  }
})
// 监听 activeTab 的变化
watch(activeTab, (newActiveTab) => {
  // 保存TAB状态
  saveTabsToLocalStorage(tabs.value, newActiveTab, appPage_storageKey)
})


// 处理“编辑”按钮点击事件
function handleEditClick(instName: string, instId: string) {
  const tabName = `编辑[${instName}]`
  const result = addTab(tabs.value, activeTab.value, tabName, 'appEditor', {
    apiUrl: `/api/v1/getApp`,
  }, {
    instName,
    instId,
    isRunning: false,
  })
  tabs.value = result.tabs
  activeTab.value = result.activeTab

  // 保存TAB状态
  saveTabsToLocalStorage(tabs.value, activeTab.value, appPage_storageKey)
}

// 处理“启动”按钮点击事件
async function handleStartClick(instName: string, instId: string, isRunning: boolean) {
  if (isRunning) {
    ElMessage.warning(`${instName} (ID: ${instId}) 已经启动`)
    return
  }

  try {
    const response = await axios.post('/api/v1/startApp', {
      instid: instId,
    })

    if (response.data.data) {
      ElMessage.success(`启动成功: ${instName} (ID: ${instId})`)
      // 更新 isRunning 状态
      const targetTab = tabs.value.find(tab => tab.jsonData.instId === instId)
      if (targetTab) {
        targetTab.jsonData.isRunning = true
      }
    }
    else {
      ElMessage.error(`启动失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`启动失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 处理“停止”按钮点击事件
async function handleStopClick(instName: string, instId: string, isRunning: boolean) {
  console.log(isRunning)
  if (!isRunning) {
    ElMessage.warning(`${instName} (ID: ${instId}) 已经停止`)
    return
  }

  try {
    const response = await axios.post('/api/v1/stopApp', {
      instid: instId,
    })

    if (response.data.data) {
      ElMessage.success(`停止成功: ${instName} (ID: ${instId})`)
      // 更新 isRunning 状态
      const targetTab = tabs.value.find(tab => tab.jsonData.instId === instId)
      if (targetTab) {
        targetTab.jsonData.isRunning = false
      }
    }
    else {
      ElMessage.error(`停止失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`停止失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 处理“删除”按钮点击事件
function handleDeleteClick(instName: string, instId: string) {
  ElMessageBox.alert(`删除：应用名称=${instName}，应用ID=${instId}`, '提示', {
    confirmButtonText: '确定',
  })
}

// 处理标签页关闭
function handleCloseTab(targetName: string) {
  const result = handleTabsEdit(tabs.value, activeTab.value, targetName, appPage_storageKey)
  tabs.value = result.tabs
  activeTab.value = result.activeTab

  // 保存TAB状态
  saveTabsToLocalStorage(tabs.value, activeTab.value, appPage_storageKey)
}
</script>

<template>
  <div>
    应用列表
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
        @edit-click="handleEditClick"
        @start-click="handleStartClick"
        @stop-click="handleStopClick"
        @delete-click="handleDeleteClick"
        @close-tab="handleCloseTab"
      />
    </el-tab-pane>
  </el-tabs>
</template>
