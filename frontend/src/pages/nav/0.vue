<script lang="ts" setup>
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref } from 'vue'
import AppEditor from './0_components/AppEditor.vue'
import AppList from './0_components/AppList.vue'

// 定义标签页类型
type TabType = 'appList' | 'appEditor'

// 定义 Tab 接口
interface Tab {
  title: string
  name: string
  type: TabType
  config: {
    apiUrl: string
  }
  jsonApp: {
    instName: string
    instId: string
    isRunning: boolean
  }
}

// 动态组件映射
const componentMap = {
  appList: AppList,
  appEditor: AppEditor,
}

// 定义 activeTab 和 tabs
const activeTab = ref('1')
const tabs = ref<Tab[]>([
  {
    title: '总览',
    name: '1',
    type: 'appList',
    config: {
      apiUrl: `/api/v1/listApps?t=${Date.now()}`,
    },
    jsonApp: {
      instName: '',
      instId: '',
      isRunning: false,
    },
  },
])

// 添加新标签页
function addTab(title: string, type: TabType, config: { apiUrl: string }, jsonApp: { instName: string, instId: string, isRunning: boolean }) {
  // 检查是否已存在相同 title 的标签页
  const existingTab = tabs.value.find(tab => tab.title === title)

  if (existingTab) {
    // 如果存在，则激活该标签页
    activeTab.value = existingTab.name
  }
  else {
    // 如果不存在，则创建新标签页
    const newTabName = `${tabs.value.length + 1}`
    const newTab = {
      title,
      name: newTabName,
      type,
      config,
      jsonApp,
    }
    tabs.value.push(newTab)
    activeTab.value = newTabName
  }
}

// 处理标签页关闭
function handleTabsEdit(targetName: string) {
  if (targetName === tabs.value[0].name) {
    console.log('最左边的标签页不允许关闭')
    return
  }
  const targetIndex = tabs.value.findIndex((tab: Tab) => tab.name === targetName)
  if (targetIndex !== -1) {
    tabs.value.splice(targetIndex, 1)
    if (activeTab.value === targetName) {
      activeTab.value = tabs.value[0]?.name || ''
    }
  }
}

// 处理“编辑”按钮点击事件
function handleEditClick(instName: string, instId: string) {
  const tabName = `编辑[${instName}]`
  addTab(tabName, 'appEditor', {
    apiUrl: `/api/v1/getApp`,
  }, {
    instName,
    instId,
    isRunning: false,
  })
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
      const targetTab = tabs.value.find(tab => tab.jsonApp.instId === instId)
      if (targetTab) {
        targetTab.jsonApp.isRunning = true
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
      const targetTab = tabs.value.find(tab => tab.jsonApp.instId === instId)
      if (targetTab) {
        targetTab.jsonApp.isRunning = false
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

// 处理关闭标签页事件
function handleCloseTab() {
  handleTabsEdit(activeTab.value)
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
    @tab-remove="handleTabsEdit"
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
        :json-app="tab.jsonApp"
        @edit-click="handleEditClick"
        @start-click="handleStartClick"
        @stop-click="handleStopClick"
        @delete-click="handleDeleteClick"
        @close-tab="handleCloseTab"
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
