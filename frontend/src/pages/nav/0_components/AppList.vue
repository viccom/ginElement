<script lang="ts" setup>
import axios from 'axios'
import { onMounted, onUnmounted, ref } from 'vue'

// 定义设备数据的类型
interface AppData {
  instName: string
  instId: string
  appCode: string
  appType: string
  autoStart: boolean
  isRunning: boolean | null // 允许 isRunning 为 null
}

const props = defineProps<{
  config: {
    apiUrl: string
  }
}>()

// 定义事件
const emit = defineEmits(['edit-click', 'start-click', 'stop-click', 'delete-click'])

// 定义 tableData 的类型为 DeviceData[]
const tableData = ref<AppData[]>([])

// 定义定时器 ID
let intervalId: number | null = null

// 获取数据的函数
async function fetchData() {
  try {
    // 发送 POST 请求
    const response = await axios.get(props.config.apiUrl, {
      headers: {
        'Content-Type': 'application/json', // 设置请求头为 JSON 格式
      },
    })

    // 定义 data 的类型
    const data: Record<string, AppData> = response.data.data

    // 将响应数据赋值给 tableData
    tableData.value = Object.values(data).map((item: AppData) => ({
      instName: item.instName,
      instId: item.instId,
      appCode: item.appCode,
      appType: item.appType,
      autoStart: item.autoStart,
      isRunning: item.isRunning,
    }))
  }
  catch (error) {
    console.error('加载表格数据失败:', error)
  }
}

// 组件挂载时启动定时器
onMounted(() => {
  // 立即获取一次数据
  fetchData()

  // 每隔 3 秒获取一次数据
  intervalId = setInterval(fetchData, 3000)
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})
</script>

<template>
  <el-table :data="tableData" style="width: 100%;">
    <el-table-column prop="instName" label="名称" />
    <el-table-column prop="instId" label="ID" />
    <el-table-column prop="appCode" label="编码" />
    <el-table-column prop="appType" label="appType" />
    <el-table-column prop="autoStart" label="自启动">
      <template #default="scope">
        <el-switch v-model="scope.row.autoStart" />
      </template>
    </el-table-column>
    <el-table-column label="状态">
      <template #default="scope">
        <el-tag
          :type="
            scope.row.isRunning === true
              ? 'success'
              : scope.row.isRunning === false
                ? 'danger'
                : 'info'
          "
        >
          {{ scope.row.isRunning === true ? '运行' : scope.row.isRunning === false ? '停止' : '未知' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="350">
      <template #default="scope">
        <el-button size="small" type="primary" @click="emit('edit-click', scope.row.instName, scope.row.instId)">
          编辑
        </el-button>
        <el-button size="small" type="success" @click="emit('start-click', scope.row.instName, scope.row.instId)">
          启动
        </el-button>
        <el-button size="small" type="warning" @click="emit('stop-click', scope.row.instName, scope.row.instId)">
          停止
        </el-button>
        <el-button size="small" type="danger" @click="emit('delete-click', scope.row.instName, scope.row.instId)">
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>
</template>
