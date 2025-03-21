<script lang="ts" setup>
import axios from 'axios'
import { ElMessage } from 'element-plus'
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
  jsonData: {
    instName: string
    instId: string
    isRunning: boolean
  }
}>()

// 定义事件
const emit = defineEmits(['edit-click', 'start-click', 'stop-click', 'delete-click', 'add-click'])

// 定义 tableData 的类型为 DeviceData[]
const tableData = ref<AppData[]>([])

// 定义定时器 ID
let intervalId: number | null = null

// 获取数据的函数
async function fetchData() {
  try {
    const response = await axios.get(props.config.apiUrl, {
      headers: {
        'Content-Type': 'application/json',
      },
    })
    const data: Record<string, AppData> = response.data.data
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

onMounted(() => {
  fetchData()
  intervalId = setInterval(fetchData, 3000)
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})

// 新增：控制删除确认对话框的显示状态
const deleteDialogVisible = ref(false)
const currentDeleteTarget = ref<{ instName: string, instId: string } | null>(null)

// 打开删除确认对话框
function openDeleteDialog(instName: string, instId: string) {
  currentDeleteTarget.value = { instName, instId }
  deleteDialogVisible.value = true
}

// 确认删除操作
async function confirmDelete() {
  if (!currentDeleteTarget.value)
    return

  try {
    const response = await axios.post('/api/v1/delApp', {
      instid: currentDeleteTarget.value.instId,
    })

    if (response.status === 200) {
      if (response.data.result === 'success') {
        ElMessage.success(`删除 ${currentDeleteTarget.value.instName} 应用成功`)
        fetchData()
      }
      else {
        ElMessage.error(`删除 ${currentDeleteTarget.value.instName} 应用失败: 应用实例绑定了设备，必须先删除绑定的设备`)
      }
      // ElMessage.success(`删除 ${currentDeleteTarget.value.instName} 应用成功`)
      // fetchData()
    }
    else {
      ElMessage.error(`删除 ${currentDeleteTarget.value.instName} 应用失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`删除 ${currentDeleteTarget.value.instName} 应用失败: ${error.response?.data?.details || '网络错误'}`)
  }
  finally {
    deleteDialogVisible.value = false
    currentDeleteTarget.value = null
  }
}

// 取消删除操作
function cancelDelete() {
  deleteDialogVisible.value = false
  currentDeleteTarget.value = null
}
</script>

<template>
  <!-- 表格部分保持不变 -->
  <el-table :data="tableData" style="width: 100%">
    <el-table-column prop="instName" label="应用名称" width="180" />
    <el-table-column prop="instId" label="应用ID" width="180" />
    <el-table-column prop="appCode" label="应用代码" width="180" />
    <el-table-column prop="appType" label="应用类型" width="180" />
    <el-table-column prop="autoStart" label="自动启动" width="180">
      <template #default="scope">
        <el-switch
          v-model="scope.row.autoStart"
          active-color="#13ce66"
          inactive-color="#ff4949"
          @change="handleAutoStartChange(scope.row)"
        />
      </template>
    </el-table-column>
    <el-table-column prop="isRunning" label="运行状态" width="180">
      <template #default="scope">
        <el-tag v-if="scope.row.isRunning === true" type="success">
          运行中
        </el-tag>
        <el-tag v-else-if="scope.row.isRunning === false" type="danger">
          已停止
        </el-tag>
        <el-tag v-else type="info">
          未知
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作">
      <template #default="scope">
        <el-button size="small" @click="emit('edit-click', scope.row.instName, scope.row.instId)">
          编辑
        </el-button>
        <el-button
          v-if="scope.row.isRunning === false"
          size="small"
          type="success"
          @click="emit('start-click', scope.row.instName, scope.row.instId, scope.row.isRunning)"
        >
          启动
        </el-button>
        <el-button
          v-if="scope.row.isRunning === true"
          size="small"
          type="warning"
          @click="emit('stop-click', scope.row.instName, scope.row.instId, scope.row.isRunning)"
        >
          停止
        </el-button>
        <el-button size="small" type="danger" @click="openDeleteDialog(scope.row.instName, scope.row.instId)">
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>

  <div style="display: flex; align-items: center; margin-bottom: 10px;">
    <!-- 新增应用按钮 -->
    <el-button type="primary" @click="emit('add-click')">
      新增应用
    </el-button>
  </div>
  <!-- 删除确认对话框 -->
  <el-dialog
    v-model="deleteDialogVisible"
    title="删除确认"
    width="30%"
    :before-close="cancelDelete"
    draggable
  >
    <span>确定要删除应用名称={{ currentDeleteTarget?.instName }}，应用ID={{ currentDeleteTarget?.instId }} 吗？</span>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="cancelDelete">取消</el-button>
        <el-button type="primary" @click="confirmDelete">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<style scoped>

</style>
