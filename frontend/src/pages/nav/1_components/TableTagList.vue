<script lang="ts" setup>
import { Refresh } from '@element-plus/icons-vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { onMounted, onUnmounted, ref } from 'vue'

interface TagDataItem {
  pointName: string
  description: string
  type: string
  // 根据实际数据扩展其他属性
}

const props = defineProps<{
  config: {
    apiUrl: string
  }
  jsonData: {
    devName: string
    devDesc: string
    devId: string
    instId: string
  }
}>()

const devName = ref(props.jsonData.devName)
const instId = ref(props.jsonData.instId)
const devDesc = ref(props.jsonData.devDesc)
const devId = ref(props.jsonData.devId)

const tableData = ref<TagDataItem[]>([])
const search = ref('') // 新增搜索功能

// 定时器ID
const intervalId: number | null = null

// 数据获取函数
async function fetchData() {
  try {
    const requestBody = {
      devList: [props.jsonData.devId],
    }
    const response = await axios.post(props.config.apiUrl, requestBody)
    const devData = response.data.data[props.jsonData.devId]
    tableData.value = Object.entries(devData).map(([_, values]) => ({
      pointName: values[0],
      description: values[1],
      type: values[2],
    }))
  }
  catch (error) {
    console.error('加载点表数据失败:', error)
  }
}

// 组件挂载时启动定时器
onMounted(() => {
  fetchData()
  // intervalId = setInterval(fetchData, 3000) // 每3秒刷新数据
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})

// 操作列按钮处理
function handleModifyClick(row: TagDataItem) {
  currentRowIndex.value = tableData.value.findIndex(item => item.pointName === row.pointName);
  currentRowData.value = { ...row };
  isDialogVisible.value = true;
}

function handleDeleteClick(pointName: string) {
  // 新增：根据pointName过滤删除当前行
  tableData.value = tableData.value.filter(item => item.pointName !== pointName)
  ElMessage({
    message: `${pointName} 已成功删除`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

async function handleImport(file: File) {
  const reader = new FileReader()
  try {
    reader.onload = (e) => {
      const text = e.target?.result as string
      const rows = text.split('\n')
      const parsedData: TagDataItem[] = []
      let errorOccurred = false

      rows.forEach((row, index) => {
        const cols = row.split(',')
        if (cols.length < 3) {
          errorOccurred = true
          ElMessage.error(`第${index + 1}行数据不足，需至少包含点名、描述、类型`)
          return
        }
        parsedData.push({
          pointName: cols[0],
          description: cols[1],
          type: cols[2],
        })
      })

      if (!errorOccurred) {
        tableData.value = parsedData
        ElMessage.success('导入成功，数据已替换')
      }
    }

    reader.onerror = (error) => {
      ElMessage.error('文件读取失败，请检查文件格式')
      console.error('文件读取错误:', error)
    }

    reader.readAsText(file)
  }
  catch (error) {
    ElMessage.error('导入过程中发生错误')
    console.error('导入错误:', error)
  }
}

function exportCSV() {
  const csvContent = tableData.value.map(({ pointName, description, type }) => `${pointName},${description},${type}`).join('\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.download = `${devName.value}-${devId.value}-tags.csv` // 新增动态文件名
  link.href = URL.createObjectURL(blob)
  link.click()
}

function saveData() {
  ElMessage({
    message: '保存功能暂未实现',
    type: 'info',
    duration: 3000,
    center: true,
  })
}

// 新增响应式变量
const isDialogVisible = ref(false);
const currentRowData = ref<TagDataItem>({ pointName: '', description: '', type: '' });
const currentRowIndex = ref(-1);

// 新增确认修改函数
function confirmModify() {
  if (currentRowIndex.value !== -1) {
    tableData.value[currentRowIndex.value] = { ...currentRowData.value };
    ElMessage.success('修改成功');
  }
  isDialogVisible.value = false;
}

</script>

<template>
  <el-row :gutter="20">
    <el-col :span="2">
      <div>
        <el-input
          v-model="search"
          placeholder="请输入点名过滤"
          clearable
        />
      </div>
    </el-col>

    <el-col :span="4">
      <div>
        <el-input
          v-model="devName"
          style="max-width: 100%"
          disabled
          placeholder="Please input"
        >
          <template #prepend>
            名称：
          </template>
        </el-input>
      </div>
    </el-col>
    <el-col :span="4">
      <div>
        <el-input
          v-model="devDesc"
          style="max-width: 100%"
          disabled
          placeholder="Please input"
        >
          <template #prepend>
            描述：
          </template>
        </el-input>
      </div>
    </el-col>
    <el-col :span="5">
      <div>
        <el-input
          v-model="devId"
          style="max-width: 100%"
          disabled
          placeholder="Please input"
        >
          <template #prepend>
            设备ID：
          </template>
        </el-input>
      </div>
    </el-col>
    <el-col :span="5">
      <div>
        <el-input
          v-model="instId"
          style="max-width: 100%"
          disabled
          placeholder="Please input"
        >
          <template #prepend>
            实例：
          </template>
        </el-input>
      </div>
    </el-col>
    <el-col :span="1">
      <div>
        <el-button type="primary" :icon="Refresh" @click="fetchData">
          重置
        </el-button>
      </div>
    </el-col>
  </el-row>

  <el-table
    :data="tableData.filter(data => !search || data.pointName.includes(search))"
    style="width: 100%"
  >
    <el-table-column prop="pointName" label="点名" />
    <el-table-column prop="description" label="描述" />
    <el-table-column prop="type" label="类型" />
    <el-table-column label="操作" width="150">
      <template #default="scope">
        <el-button
          size="small"
          type="primary"
          @click="handleModifyClick(scope.row)"
        >
          修改
        </el-button>
        <el-button
          size="small"
          type="danger"
          @click="handleDeleteClick(scope.row.pointName)"
        >
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>

  <!-- 新增按钮组 -->
  <el-row style="margin-top: 20px">
    <el-col :span="6">
      <el-upload
        action="#"
        :show-file-list="false"
        :on-change="(uploadFile) => handleImport(uploadFile.raw)"
        accept=".csv"
      >
        <el-button type="primary">
          导入
        </el-button>
      </el-upload>
    </el-col>
    <el-col :span="6">
      <el-button type="success" @click="exportCSV">
        导出
      </el-button>
    </el-col>
    <el-col :span="6">
      <el-button type="warning" @click="saveData">
        保存
      </el-button>
    </el-col>
  </el-row>

  <!-- 新增对话框组件 -->
  <el-dialog
    v-model="isDialogVisible"
    title="修改数据"
    width="30%"
  >
    <el-form :model="currentRowData">
      <el-form-item label="点名">
        <el-input v-model="currentRowData.pointName" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="currentRowData.description" />
      </el-form-item>
      <el-form-item label="类型">
        <el-input v-model="currentRowData.type" />
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="isDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmModify">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>
