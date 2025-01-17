<script lang="ts" setup>
import { ElMessageBox } from 'element-plus'
import { ref } from 'vue'
import ChartComponent from './1_components/ChartComponent.vue'
import FormComponent from './1_components/FormComponent.vue'
import TableDataList from './1_components/TableDataList.vue'
import TableDevList from './1_components/TableDevList.vue'

// 定义标签页类型
type TabType = 'devs' | 'form' | 'data' | 'chart'

// 定义 Tab 接口
interface Tab {
  title: string
  name: string
  type: TabType
  config: {
    apiUrl: string
  }
  jsonDev: {
    devName: string
    devDesc: string
    devId: string
    instId: string
  }
}

// 动态组件映射
const componentMap = {
  devs: TableDevList,
  data: TableDataList,
  form: FormComponent,
  chart: ChartComponent,
}

// 定义 activeTab 和 tabs
const activeTab = ref('1')
const tabs = ref<Tab[]>([
  {
    title: '设备列表',
    name: '1',
    type: 'devs',
    config: {
      apiUrl: '/api/v1/listDevices',
    },
    jsonDev: {
      devName: '',
      devDesc: '',
      devId: '',
      instId: '',
    },
  },
])

// 添加新标签页
function addTab(title: string, type: TabType, config: { apiUrl: string }, jsonDev: { devName: string, devDesc: string, devId: string, instId: string }) {
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
      jsonDev,
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

// 处理“数据”按钮点击事件
function handleDataClick(devName: string, devDesc: string, devId: string, instId: string) {
  const tabName = `数据[${devName}]`
  addTab(tabName, 'data', {
    apiUrl: `/api/v1/getDevvalues`,
  }, {
    devName,
    devDesc,
    devId,
    instId,
  })
}

// 处理“点表”按钮点击事件
function handlePointClick(devName: string, devId: string) {
  ElMessageBox.alert(`设备点表：<br>设备名称=${devName}<br>设备ID=${devId}`, '提示', {
    confirmButtonText: '确定',
    dangerouslyUseHTMLString: true, // 允许使用 HTML 字符串
  })
}

// 处理“删除”按钮点击事件
function handleDeleteClick(devName: string, devId: string) {
  ElMessageBox.alert(`删除设备：<br>设备名称=${devName}<br>设备ID=${devId}`, '提示', {
    confirmButtonText: '确定',
    dangerouslyUseHTMLString: true, // 允许使用 HTML 字符串
  })
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
        :json-dev="tab.jsonDev"
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
